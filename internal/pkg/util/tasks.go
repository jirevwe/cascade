package util

import (
	"time"
)

type TaskName string

type QueueName string

const DeleteEntityTask TaskName = "DeleteEntityTask"

const DeleteEntityQueue QueueName = "delete_entity_queue"

type QueueError struct {
	delay time.Duration
	Err   error
}

func (e *QueueError) Error() string {
	return e.Err.Error()
}

func (e *QueueError) Delay() time.Duration {
	return e.delay
}

type RateLimitError struct {
	delay time.Duration
	Err   error
}

func (e *RateLimitError) Error() string {
	return e.Err.Error()
}

func (e *RateLimitError) Delay() time.Duration {
	return e.delay
}
