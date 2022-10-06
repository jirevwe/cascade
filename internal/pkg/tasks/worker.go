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
		var child config.Entity
		err := json.Unmarshal(t.Payload(), &child)
		if err != nil {
			logrus.Errorf("json: an error occured unmarshaling payload - %+v", err)
			return err
		}

		cmd := rdb.Client().Get(ctx, fmt.Sprintf("%s:%s:filter", child.Name, child.Id))
		filterStr, err := cmd.Result()
		if err != nil {
			logrus.Errorf("redis: an error occured getting filter - %+v", err)
			return err
		}

		cmd = rdb.Client().Get(ctx, fmt.Sprintf("%s:%s:update", child.Name, child.Id))
		updateStr, err := cmd.Result()
		if err != nil {
			logrus.Errorf("redis: an error occured getting update - %+v", err)
			return err
		}

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

		ctx = context.WithValue(ctx, datastore.CollectionCtx, child.Name)
		err = store.UpdateMany(ctx, filter, update, true)
		if err != nil {
			logrus.Errorf("mongodb: an error occured updating %s - %+v", child.Name, err)
			return err
		}

		logrus.Info("asynq: soft-deleted children records")

		return nil
	}
}
