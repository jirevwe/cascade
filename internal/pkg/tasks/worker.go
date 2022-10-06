package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/jirevwe/cascade/internal/pkg/config"
	"github.com/jirevwe/cascade/internal/pkg/datastore"
	"github.com/jirevwe/cascade/internal/pkg/queue"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Worker interface {
	Start()
	Stop()
}

func DeleteEntity(store datastore.DB, rdb *queue.Redis) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		id := string(t.Payload())

		cmd := rdb.Client().Get(ctx, fmt.Sprintf("%s:filter", id))
		filterStr, err := cmd.Result()
		if err != nil {
			logrus.Errorf("redis: an error occured getting filter - %+v", err)
			return err
		}

		cmd = rdb.Client().Get(ctx, fmt.Sprintf("%s:update", id))
		updateStr, err := cmd.Result()
		if err != nil {
			logrus.Errorf("redis: an error occured getting update - %+v", err)
			return err
		}

		cmd = rdb.Client().Get(ctx, fmt.Sprintf("%s:relation", id))
		relationStr, err := cmd.Result()
		if err != nil {
			logrus.Errorf("redis: an error occured getting relation - %+v", err)
			return err
		}

		logrus.Info("filter: ", filterStr)
		logrus.Info("update: ", updateStr)
		logrus.Info("relation: ", relationStr)

		var filter primitive.M
		err = json.Unmarshal([]byte(filterStr), &filter)
		if err != nil {
			logrus.Errorf("json: an error occured unmarshaling filter - %+v", err)
			return err
		}

		var update primitive.M
		err = json.Unmarshal([]byte(updateStr), &update)
		if err != nil {
			logrus.Errorf("json: an error occured unmarshaling update - %+v", err)
			return err
		}

		var relation config.Entity
		err = json.Unmarshal([]byte(relationStr), &relation)
		if err != nil {
			logrus.Errorf("json: an error occured unmarshaling relation - %+v", err)
			return err
		}

		ctx = context.WithValue(ctx, datastore.CollectionCtx, relation.Name)
		err = store.UpdateMany(ctx, filter, update, true)
		if err != nil {
			logrus.Errorf("mongodb: an error occured updating %s - %+v", relation.Name, err)
			return err
		}

		logrus.Info("asynq: soft-deleted children records")

		return nil
	}
}
