package main

import (
	"context"
)

type CollectorProcessorQueryInterface interface {
	Setup(collector *CollectorQuery)
	Reset()
	Collect(ctx context.Context, callback chan<- func())
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

	return &collector
}
