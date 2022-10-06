package tasks

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jirevwe/cascade/internal/pkg/queue"
	"github.com/jirevwe/cascade/internal/pkg/util"
	log "github.com/sirupsen/logrus"
)

type Consumer struct {
	queue queue.Queuer
	mux   *asynq.ServeMux
	srv   *asynq.Server
}

func NewConsumer(q queue.Queuer) (*Consumer, error) {
	srv := asynq.NewServer(
		q.Options().RedisClient,
		asynq.Config{
			Concurrency: 10,
			Queues:      q.Options().Names,
			IsFailure: func(err error) bool {
				// TODO(raymond): check for error type here
				return true
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Second * 10
			},
		},
	)

	mux := asynq.NewServeMux()
	c := &Consumer{queue: q, mux: mux, srv: srv}

	return c, nil
}

func (c *Consumer) Start() {
	if err := c.srv.Start(c.mux); err != nil {
		log.WithError(err).Fatal("error starting worker")
	}
}

func (c *Consumer) RegisterHandlers(taskName util.TaskName, handler func(context.Context, *asynq.Task) error) {
	c.mux.HandleFunc(string(taskName), handler)
}

func (c *Consumer) Stop() {
	c.srv.Stop()
	c.srv.Shutdown()
}
