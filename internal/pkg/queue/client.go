package queue

import (
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/jaevor/go-nanoid"
	"github.com/jirevwe/cascade/internal/pkg/config"
	"github.com/jirevwe/cascade/internal/pkg/util"
	"github.com/sirupsen/logrus"
)

type RedisQueue struct {
	opts      QueueOptions
	client    *asynq.Client
	inspector *asynq.Inspector
}

func NewClient(cfg config.Configuration) (*asynq.Client, error) {
	rdb, err := NewRedis(cfg.RedisDsn)
	if err != nil {
		return nil, err
	}

	client := asynq.NewClient(rdb)

	return client, nil
}

func NewQueue(opts QueueOptions) Queuer {
	client := asynq.NewClient(opts.RedisClient)
	inspector := asynq.NewInspector(opts.RedisClient)
	return &RedisQueue{
		client:    client,
		opts:      opts,
		inspector: inspector,
	}
}

func (q *RedisQueue) Write(taskName util.TaskName, queueName util.QueueName, job *Job) error {
	if job.ID == "" {
		generateID, err := nanoid.Standard(21)
		if err != nil {
			return err
		}

		job.ID = generateID()
	}

	t := asynq.NewTask(string(taskName), job.Payload, asynq.Queue(string(queueName)), asynq.TaskID(job.ID), asynq.ProcessIn(job.Delay))
	info, err := q.client.Enqueue(t)
	if err != nil {
		return err
	}

	logrus.Infof("asynq: %+v", info)

	return nil
}

func (q *RedisQueue) Options() QueueOptions {
	return q.opts
}

func (q *RedisQueue) Monitor() *asynqmon.HTTPHandler {
	h := asynqmon.New(asynqmon.Options{
		RootPath:          "/queue/monitoring",
		RedisConnOpt:      q.opts.RedisClient,
		PrometheusAddress: q.opts.PrometheusAddress,
	})
	return h
}

func (q *RedisQueue) Inspector() *asynq.Inspector {
	return q.inspector
}
