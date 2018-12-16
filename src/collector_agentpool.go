package main

import (
	"context"
	"sync"
	"time"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
)

type CollectorAgentPool struct {
	Processor CollectorProcessorAgentPoolInterface
	Name string
	ScrapeTime  *time.Duration
	AzureDevOpsProjects *devopsClient.ProjectList
	AgentPoolIdList []int64
}

func (m *CollectorAgentPool) Run(scrapeTime time.Duration) {
	m.ScrapeTime = &scrapeTime

	m.Processor.Setup(m)
	go func() {
		for {
			go func() {
				m.Collect()
			}()
			Logger.Verbose("collector[%s]: sleeping %v", m.Name, m.ScrapeTime.String())
			time.Sleep(*m.ScrapeTime)
		}
	}()
}

func (m *CollectorAgentPool) Collect() {
	var wg sync.WaitGroup
	var wgCallback sync.WaitGroup

	ctx := context.Background()

	callbackChannel := make(chan func())

	Logger.Messsage(
		"collector[%s]: starting metrics collection",
		m.Name,
	)

	wg.Add(1)
	go func(ctx context.Context, callback chan<- func()) {
		defer wg.Done()
		m.Processor.Collect(ctx, callbackChannel)
	}(ctx, callbackChannel)

	// collect metrics (callbacks) and proceses them
	wgCallback.Add(1)
	go func() {
		defer wgCallback.Done()
		var callbackList []func()
		for callback := range callbackChannel {
			callbackList = append(callbackList, callback)
		}

		// reset metric values
		m.Processor.Reset()

		// process callbacks (set metrics)
		for _, callback := range callbackList {
			callback()
		}
	}()

	// wait for all funcs
	wg.Wait()
	close(callbackChannel)
	wgCallback.Wait()

	Logger.Verbose(
		"collector[%s]: finished metrics collection",
		m.Name,
	)
}
