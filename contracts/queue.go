package contracts

import "time"

type QueueName string
type Queue interface {
	NewQueue(name QueueName, handler func(string) bool)
	PushRaw(queueName QueueName, payload string) error
	Push(queueName QueueName, payloadPtr any) error
	PushDelayed(queueName QueueName, payloadPtr any, delayDuration time.Duration) error
	PushDelayedRaw(queueName QueueName, payload string, delayDuration time.Duration) error
	PushScheduledRaw(queueName QueueName, payload string, t time.Time) error
	PushScheduled(queueName QueueName, payloadPtr any, t time.Time) error
}
