package contracts

import "github.com/go-co-op/gocron/v2"

type CronJob interface {
	DefineJob(definition gocron.JobDefinition, handler func()) error
}
