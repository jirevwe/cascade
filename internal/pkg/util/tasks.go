package util

import (
	"time"
)

type TaskName string

type QueueName string

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
