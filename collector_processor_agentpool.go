package main

import (
	"context"
	log "github.com/sirupsen/logrus"
)

type CollectorProcessorAgentPoolInterface interface {
	Setup(collector *CollectorAgentPool)
	Reset()
	Collect(ctx context.Context, contextLogger *log.Entry, callback chan<- func())
}

type CollectorProcessorAgentPool struct {
	CollectorProcessorAgentPoolInterface
	CollectorReference *CollectorAgentPool
}

func NewCollectorAgentPool(name string, processor CollectorProcessorAgentPoolInterface) *CollectorAgentPool {
	collector := CollectorAgentPool{
		CollectorBase: CollectorBase{
			Name: name,
		},
		Processor: processor,
	}
	collector.CollectorBase.Init()

	return &collector
}

func (c *CollectorProcessorAgentPool) logger() *log.Entry {
	return c.CollectorReference.logger
}
