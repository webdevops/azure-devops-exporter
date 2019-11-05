package main

import (
	"context"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type CollectorProcessorProjectInterface interface {
	Setup(collector *CollectorProject)
	Reset()
	Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project)
}

type CollectorProcessorProject struct {
	CollectorProcessorProjectInterface
	CollectorReference *CollectorProject
}

func NewCollectorProject(name string, processor CollectorProcessorProjectInterface) *CollectorProject {
	collector := CollectorProject{
		CollectorBase: CollectorBase{
			Name: name,
		},
		Processor: processor,
	}

	return &collector
}
