package main

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type CollectorProcessorGeneralInterface interface {
	Setup(collector *CollectorGeneral)
	Reset()
	Collect(ctx context.Context, contextLogger *log.Entry, callback chan<- func())
}

type CollectorProcessorGeneral struct {
	CollectorProcessorGeneralInterface
	CollectorReference *CollectorGeneral
}

func NewCollectorGeneral(name string, processor CollectorProcessorGeneralInterface) *CollectorGeneral {
	collector := CollectorGeneral{
		CollectorBase: CollectorBase{
			Name: name,
		},
		Processor: processor,
	}
	collector.CollectorBase.Init()

	return &collector
}
