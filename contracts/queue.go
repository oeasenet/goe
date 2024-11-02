package contracts

import "time"

type QueueName string
type Queue interface {
	NewQueue(name QueueName, handler func(string) bool, cfgs ...*NewQueueCfg) error
	PushRaw(queueName QueueName, payload string) error
	Push(queueName QueueName, payloadPtr any) error
	PushDelayed(queueName QueueName, payloadPtr any, delayDuration time.Duration) error
	PushDelayedRaw(queueName QueueName, payload string, delayDuration time.Duration) error
	PushScheduledRaw(queueName QueueName, payload string, t time.Time) error
	PushScheduled(queueName QueueName, payloadPtr any, t time.Time) error
}

// NewQueueCfg is the configuration for creating a new queue
type NewQueueCfg struct {
	// ConcurrentWorkers is the number of concurrent workers to process the queue
	ConcurrentWorkers int
	// FetchInterval is the interval in seconds to fetch new jobs
	FetchInterval int
	// DefaultRetries is number of retries for a job, can be overridden when creating a message
	DefaultRetries int
	// MaxConsumeDuration is the maximum time in seconds to consume a job, if exceeded, the job will be retried
	MaxConsumeDuration int
	// FetchLimit is the maximum number of jobs to fetch in a single fetch, 0 means no limit
	FetchLimit int
}
