package cron

import (
	"errors"
	"github.com/go-co-op/gocron/v2"
)

type CronJobModule struct {
	scheduler gocron.Scheduler
	started   bool
}

func NewCronJobService() (*CronJobModule, error) {
	s, err := gocron.NewScheduler(gocron.WithLimitConcurrentJobs(1, gocron.LimitModeWait))
	if err != nil {
		return nil, err
	}
	return &CronJobModule{
		scheduler: s,
	}, nil
}

func (c *CronJobModule) Start() {
	if c.scheduler == nil {
		return
	}
	if len(c.scheduler.Jobs()) == 0 {
		return
	}
	if c.started {
		return
	}
	c.scheduler.Start()
	c.started = true
}

func (c *CronJobModule) Close() error {
	if c.scheduler == nil {
		return errors.New("scheduler is not initialized")
	}
	if !c.started {
		return errors.New("scheduler is not started")
	}
	return c.scheduler.Shutdown()
}

func (c *CronJobModule) DefineJob(definition gocron.JobDefinition, handler func()) error {
	if c.scheduler == nil {
		return errors.New("scheduler is not initialized")
	}
	if c.started {
		return errors.New("scheduler is already started, jobs can only be defined before starting the scheduler")
	}
	_, err := c.scheduler.NewJob(definition, gocron.NewTask(handler))
	return err
}
