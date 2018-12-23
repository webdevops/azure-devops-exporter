package main

import (
	"context"
)

type CollectorProcessorAgentPoolInterface interface {
	Setup(collector *CollectorAgentPool)
	Reset()
	Collect(ctx context.Context, callback chan<- func())
}

type CollectorProcessorAgentPool struct {
	CollectorProcessorAgentPoolInterface
	CollectorReference *CollectorAgentPool
}

func NewCollectorAgentPool(name string, processor CollectorProcessorAgentPoolInterface) *CollectorAgentPool {
	collector := CollectorAgentPool{
		CollectorBase: CollectorBase{
			Name:      name,
		},
		Processor: processor,
	}

	return &collector
}
