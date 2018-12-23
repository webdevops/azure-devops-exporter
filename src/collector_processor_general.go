package main

import (
	"context"
)

type CollectorProcessorGeneralInterface interface {
	Setup(collector *CollectorGeneral)
	Reset()
	Collect(ctx context.Context, callback chan<- func())
}

type CollectorProcessorGeneral struct {
	CollectorProcessorGeneralInterface
	CollectorReference *CollectorGeneral
}

func NewCollectorGeneral(name string, processor CollectorProcessorGeneralInterface) *CollectorGeneral {
	collector := CollectorGeneral{
		CollectorBase: CollectorBase{
			Name:      name,
		},
		Processor: processor,
	}

	return &collector
}
