package core

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/queue"
	"sync"
	"time"
)

type GoeQueue struct {
	queues    sync.Map
	goeConfig *GoeConfig
	logger    contracts.Logger
	redisCli  *redis.Client
}

func NewGoeQueue(appConfig *GoeConfig, logger contracts.Logger) (*GoeQueue, error) {
	if appConfig.Redis.Host != "" && appConfig.Redis.Port != 0 {
		redisCli := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", appConfig.Redis.Host, appConfig.Redis.Port),
			Username: appConfig.Redis.Username,
			Password: appConfig.Redis.Password,
			DB:       RedisDBMQ,
		})
		return &GoeQueue{
			goeConfig: appConfig,
			logger:    logger,
			queues:    sync.Map{},
			redisCli:  redisCli,
		}, nil
	} else {
		return nil, errors.New("failed to initialize Redis MQ: missing required redis configuration")
	}
}

func (g *GoeQueue) Start() error {
	g.queues.Range(func(key, value any) bool {
		rq, ok := value.(*queue.DelayQueue)
		if !ok {
			return false
		}
		rq.StartConsume()
		return true
	})
	return nil
}

func (g *GoeQueue) Close() error {
	g.queues.Range(func(key, value any) bool {
		rq, ok := value.(*queue.DelayQueue)
		if !ok {
			return false
		}
		rq.StopConsume()
		return true
	})
	return nil
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

func (g *GoeQueue) NewQueue(name contracts.QueueName, handler func(string) bool, cfgs ...*NewQueueCfg) {
	rq := queue.NewQueue(string(name), g.redisCli)
	rq.WithCallback(handler)
	if len(cfgs) == 0 && cfgs[0] != nil {
		// if config is provided, use the provided config, if values are 0, use the default values
		if cfgs[0].ConcurrentWorkers == 0 {
			// default to 1 worker
			cfgs[0].ConcurrentWorkers = 1
		}
		if cfgs[0].FetchInterval == 0 {
			// default to 5 seconds
			cfgs[0].FetchInterval = 5
		}
		if cfgs[0].DefaultRetries == 0 {
			// default to 3 retries
			cfgs[0].DefaultRetries = 3
		}
		if cfgs[0].MaxConsumeDuration == 0 {
			// default to 60 seconds
			cfgs[0].MaxConsumeDuration = 60
		}
		rq.WithConcurrent(uint(cfgs[0].ConcurrentWorkers))
		rq.WithFetchInterval(time.Duration(cfgs[0].FetchInterval) * time.Second)
		rq.WithDefaultRetryCount(uint(cfgs[0].DefaultRetries))
		rq.WithMaxConsumeDuration(time.Duration(cfgs[0].MaxConsumeDuration) * time.Second)
		rq.WithFetchLimit(uint(cfgs[0].FetchLimit))
	} else {
		// if no config is provided, use the default config from the app config
		rq.WithConcurrent(uint(g.goeConfig.Queue.ConcurrentWorkers))
		rq.WithFetchInterval(time.Duration(g.goeConfig.Queue.FetchInterval) * time.Second)
		rq.WithDefaultRetryCount(uint(g.goeConfig.Queue.DefaultRetries))
		rq.WithMaxConsumeDuration(time.Duration(g.goeConfig.Queue.MaxConsumeDuration) * time.Second)
		rq.WithFetchLimit(uint(g.goeConfig.Queue.FetchLimit))
	}
	g.queues.Store(name, rq)
}

func (g *GoeQueue) PushRaw(queueName contracts.QueueName, payload string) error {
	rqm, ok := g.queues.Load(queueName)
	if !ok {
		return errors.New("queue not found")
	}
	rq, ok := rqm.(*queue.DelayQueue)
	if !ok {
		return errors.New("queue not found or invalid type")
	}
	return rq.SendScheduleMsg(payload, time.Now())
}

func (g *GoeQueue) Push(queueName contracts.QueueName, payloadPtr any) error {
	rqm, ok := g.queues.Load(queueName)
	if !ok {
		return errors.New("queue not found")
	}
	rq, ok := rqm.(*queue.DelayQueue)
	if !ok {
		return errors.New("queue not found or invalid type")
	}
	data, err := json.Marshal(payloadPtr)
	if err != nil {
		return err
	}
	return rq.SendScheduleMsg(string(data), time.Now())
}

func (g *GoeQueue) PushDelayed(queueName contracts.QueueName, payloadPtr any, delayDuration time.Duration) error {
	rqm, ok := g.queues.Load(queueName)
	if !ok {
		return errors.New("queue not found")
	}
	rq, ok := rqm.(*queue.DelayQueue)
	if !ok {
		return errors.New("queue not found or invalid type")
	}
	data, err := json.Marshal(payloadPtr)
	if err != nil {
		return err
	}
	return rq.SendDelayMsg(string(data), delayDuration)
}

func (g *GoeQueue) PushDelayedRaw(queueName contracts.QueueName, payload string, delayDuration time.Duration) error {
	rqm, ok := g.queues.Load(queueName)
	if !ok {
		return errors.New("queue not found")
	}
	rq, ok := rqm.(*queue.DelayQueue)
	if !ok {
		return errors.New("queue not found or invalid type")
	}
	return rq.SendDelayMsg(payload, delayDuration)
}

func (g *GoeQueue) PushScheduledRaw(queueName contracts.QueueName, payload string, t time.Time) error {
	rqm, ok := g.queues.Load(queueName)
	if !ok {
		return errors.New("queue not found")
	}
	rq, ok := rqm.(*queue.DelayQueue)
	if !ok {
		return errors.New("queue not found or invalid type")
	}
	return rq.SendScheduleMsg(payload, t)
}

func (g *GoeQueue) PushScheduled(queueName contracts.QueueName, payloadPtr any, t time.Time) error {
	rqm, ok := g.queues.Load(queueName)
	if !ok {
		return errors.New("queue not found")
	}
	rq, ok := rqm.(*queue.DelayQueue)
	if !ok {
		return errors.New("queue not found or invalid type")
	}
	data, err := json.Marshal(payloadPtr)
	if err != nil {
		return err
	}
	return rq.SendScheduleMsg(string(data), t)
}
