package main

import (
	"context"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
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

func NewCollectorProject(name string, processor CollectorProcessorProjectInterface) (*CollectorProject) {
	collector := CollectorProject{
		Name: name,
		Processor: processor,
	}

	return &collector
}
