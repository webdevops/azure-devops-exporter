package main

import (
	"context"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"sync"
	"time"
)

type CollectorGeneral struct {
	Processor CollectorProcessorGeneralInterface
	Name string
	ScrapeTime  *time.Duration
	AzureDevOpsProjects *devopsClient.ProjectList
}

func (m *CollectorGeneral) Run(scrapeTime time.Duration) {
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

func (m *CollectorGeneral) Collect() {
	var wg sync.WaitGroup
	var wgCallback sync.WaitGroup

	if m.AzureDevOpsProjects == nil {
		Logger.Messsage(
			"collector[%s]: no projects found, skipping",
			m.Name,
		)
		return
	}

	ctx := context.Background()

	callbackChannel := make(chan func())

	Logger.Messsage(
		"collector[%s]: starting metrics collection",
		m.Name,
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Processor.Collect(ctx, callbackChannel)
	}()

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
