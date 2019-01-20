package main

import (
	"context"
	"sync"
	"time"
)

type CollectorGeneral struct {
	CollectorBase

	Processor CollectorProcessorGeneralInterface
}

func (c *CollectorGeneral) Run(scrapeTime time.Duration) {
	c.SetScrapeTime(scrapeTime)

	c.Processor.Setup(c)
	go func() {
		for {
			go func() {
				c.Collect()
			}()
			c.sleepUntilNextCollection()
		}
	}()
}

func (c *CollectorGeneral) Collect() {
	var wg sync.WaitGroup
	var wgCallback sync.WaitGroup

	if c.GetAzureProjects() == nil {
		Logger.Infof("collector[%s]: no projects found, skipping", c.Name)
		return
	}

	ctx := context.Background()

	callbackChannel := make(chan func())

	c.collectionStart()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Processor.Collect(ctx, callbackChannel)
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
		c.Processor.Reset()

		// process callbacks (set metrics)
		for _, callback := range callbackList {
			callback()
		}
	}()

	// wait for all funcs
	wg.Wait()
	close(callbackChannel)
	wgCallback.Wait()

	c.collectionFinish()
}

