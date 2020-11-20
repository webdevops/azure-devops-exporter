package main

import (
	"context"
	log "github.com/sirupsen/logrus"
)

type CollectorProcessorQueryInterface interface {
	Setup(collector *CollectorQuery)
	Reset()
	Collect(ctx context.Context, contextLogger *log.Entry, callback chan<- func())
}

type CollectorProcessorQuery struct {
	CollectorProcessorQueryInterface
	CollectorReference *CollectorQuery
}

func NewCollectorQuery(name string, processor CollectorProcessorQueryInterface) *CollectorQuery {
	collector := CollectorQuery{
		CollectorBase: CollectorBase{
			Name: name,
		},
		Processor: processor,
	}
	collector.CollectorBase.Init()

	return &collector
}
