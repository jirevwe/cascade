package listen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jirevwe/cascade/internal/pkg/config"
	"github.com/jirevwe/cascade/internal/pkg/datastore"
	"github.com/jirevwe/cascade/internal/pkg/queue"
	"github.com/jirevwe/cascade/internal/pkg/util"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	GenericMap    map[string]interface{}
	OperationType string
)

type Entity struct {
	ID        primitive.ObjectID `bson:"_id"`
	DeletedAt primitive.DateTime `bson:"deleted_at"`
}

func (o OperationType) String() string {
	return string(o)
}

const (
	ReplaceOp OperationType = "replace"
	DeleteOp  OperationType = "delete"
)

func New(cfg config.Configuration, db datastore.DB, rdb *queue.Redis, queue queue.Queuer) {
	for _, relation := range cfg.Relations {
		go listenToChangeStream(db, rdb, queue, relation)
	}
}

func listenToChangeStream(store datastore.DB, rdb *queue.Redis, q queue.Queuer, relation config.Relation) {
	ctx := context.Background()
	db := store.GetDatabase()
	coll := db.Collection(relation.Dest.Name)

	cs, err := coll.Watch(ctx, mongo.Pipeline{bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: relation.On}},
		},
	}})
	if err != nil {
		log.Fatal(err)
	}

	defer cs.Close(ctx)

	log.Printf("listen: started listener (db: %s, collection: %s)", db.Name(), coll.Name())

	for {
		ok := cs.Next(ctx)
		if ok {
			var document GenericMap
			err := cs.Decode(&document)
			if err != nil {
				log.Fatal(err)
			}

			if v, ok := document["operationType"]; ok {
				operation, ok := v.(string)
				if !ok {
					continue
				}

				if operation == relation.On {
					if vv, ok := document["fullDocument"]; ok {
						doc, ok := vv.(GenericMap)
						if !ok {
							continue
						}

						log.Printf("doc: %+v", doc)
						if vvv, ok := doc["deleted_at"]; ok {
							deletedAt, ok := vvv.(primitive.DateTime)
							if !ok {
								continue
							}

							id := doc["_id"].(primitive.ObjectID).Hex()
							filter := bson.M{relation.Source.PrimaryKey: doc[relation.Dest.ForeignKey].(string)}
							update := bson.M{"$set": bson.M{"deleted_at": deletedAt}}

							// marshall into bytes
							filterBytes, err := json.Marshal(filter)
							if err != nil {
								log.Errorf("json: could not marshal filter map - %+v", err)
								continue
							}

							updateBytes, err := json.Marshal(update)
							if err != nil {
								log.Errorf("json: could not marshal filter map - %+v", err)
								continue
							}

							relationBytes, err := json.Marshal(relation)
							if err != nil {
								log.Errorf("json: could not marshal filter map - %+v", err)
								continue
							}

							// write to redis
							cmd := rdb.Client().Set(ctx, fmt.Sprintf("%s:filter", id), filterBytes, time.Hour)
							if cmd.Err() != nil {
								log.Errorf("redis: could not write filter - %+v", cmd.Err())
								continue
							}

							cmd = rdb.Client().Set(ctx, fmt.Sprintf("%s:update", id), updateBytes, time.Hour)
							if cmd.Err() != nil {
								log.Errorf("redis: could not write update - %+v", cmd.Err())
								continue
							}

							cmd = rdb.Client().Set(ctx, fmt.Sprintf("%s:relation", id), relationBytes, time.Hour)
							if cmd.Err() != nil {
								log.Errorf("redis: could not write relation - %+v", cmd.Err())
								continue
							}

							// write to queue
							job := queue.Job{
								ID:      id,
								Payload: json.RawMessage(id),
								Delay:   1 * time.Second,
							}

							err = q.Write(util.DeleteEntityTask, util.DeleteEntityQueue, &job)
							if err != nil {
								log.Errorf("asynq: could not write to the queue - %+v", err)
								continue
							}

							logrus.Infof("listen: added records for %s to the queue", id)
						}
					}
				}
			}
		}
	}
}
