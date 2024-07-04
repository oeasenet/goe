package core

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/queue"
	"go.oease.dev/goe/utils"
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
	if appConfig.Redis.Host != "" && appConfig.Redis.Port != 0 && appConfig.Redis.Username != "" && appConfig.Redis.Password != "" {
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

func (g *GoeQueue) NewQueue(name contracts.QueueName, handler func(string) bool) {
	rq := queue.NewQueue(string(name), g.redisCli)
	rq.WithCallback(handler)
	rq.WithConcurrent(uint(g.goeConfig.Queue.ConcurrentWorkers))
	rq.WithFetchInterval(time.Duration(g.goeConfig.Queue.FetchInterval) * time.Second)
	rq.WithDefaultRetryCount(uint(g.goeConfig.Queue.DefaultRetries))
	rq.WithMaxConsumeDuration(time.Duration(g.goeConfig.Queue.MaxConsumeDuration) * time.Second)
	rq.WithFetchLimit(uint(g.goeConfig.Queue.FetchLimit))
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
	if utils.CheckIfPointer(payloadPtr) {
		return errors.New("payload must be a pointer")
	}
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
	if utils.CheckIfPointer(payloadPtr) {
		return errors.New("payload must be a pointer")
	}
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
	if utils.CheckIfPointer(payloadPtr) {
		return errors.New("payload must be a pointer")
	}
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
