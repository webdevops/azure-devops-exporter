package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type CollectorProcessorProjectInterface interface {
	Setup(collector *CollectorProject)
	Reset()
	Collect(ctx context.Context, contextLogger *log.Entry, callback chan<- func(), project devopsClient.Project)
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
	collector.CollectorBase.Init()

	return &collector
}
