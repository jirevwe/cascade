package queue

import (
	"errors"

	"github.com/go-redis/redis/v8"
)

// Redis is our wrapper logic to instrument redis calls
type Redis struct {
	dsn    string
	client *redis.Client
}

// NewClient is used to create new Redis type. This type
// encapsulates our interaction with redis and provides instrumentation with new relic.
func NewRedis(dsn string) (*Redis, error) {
	if len(dsn) == 0 {
		return nil, errors.New("redis dsn cannot be empty")
	}

	opts, err := redis.ParseURL(dsn)

	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	return &Redis{dsn: dsn, client: client}, nil
}

// Client is to return underlying redis interface
func (r *Redis) Client() *redis.Client {
	return r.client
}

// MakeRedisClient is used to fulfill asynq's interface
func (r *Redis) MakeRedisClient() interface{} {
	return r.client
}
