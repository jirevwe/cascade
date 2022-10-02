package listen

import (
	"context"
	"time"

	"github.com/jirevwe/cascade/internal/pkg/config"
	"github.com/jirevwe/cascade/internal/pkg/datastore"
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

func New(cfg config.Configuration, db datastore.DB) {
	for _, relation := range cfg.Relations {
		go listenToChangeStream(db, relation)
	}
}

func listenToChangeStream(store datastore.DB, relation config.Relation) {
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
							_, ok := vvv.(primitive.DateTime)
							if !ok {
								continue
							}

							filter := bson.M{relation.Source.PrimaryKey: doc[relation.Dest.ForeignKey].(string)}
							update := bson.M{"$set": bson.M{"deleted_at": primitive.NewDateTimeFromTime(time.Now())}}

							ctx = context.WithValue(ctx, datastore.CollectionCtx, relation.Source.Name)
							err := store.UpdateMany(ctx, filter, update, true)
							if err != nil {
								log.Errorf("an error occured updating %s - %+v", relation.Source.Name, err)
								continue
							}

							log.Info("soft-deleted children docs")
						}
					}
				}
			}
		}
	}
}
