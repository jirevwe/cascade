package queue

import (
	"encoding/json"
	"time"

	"github.com/jirevwe/cascade/internal/pkg/util"
)

type Queuer interface {
	Write(util.TaskName, util.QueueName, *Job) error
	Options() QueueOptions
}

type Job struct {
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
	Delay   time.Duration   `json:"delay"`
}

type QueueOptions struct {
	Names             map[string]int
	Type              string
	RedisClient       *Redis
	RedisAddress      string
	PrometheusAddress string
}
