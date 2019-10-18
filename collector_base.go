package main

import (
	"sync"
	"time"

	devopsClient "github.com/keremispirli/azure-devops-exporter/tree/addQueryMetrics/azure-devops-client"
)

type CollectorBase struct {
	Name       string
	scrapeTime *time.Duration

	azureDevOpsProjects      *devopsClient.ProjectList
	azureDevOpsProjectsMutex sync.Mutex

	LastScrapeDuration  *time.Duration
	collectionStartTime *time.Time
	collectionLastTime  *time.Time
}

func (c *CollectorBase) Init() {
}

func (c *CollectorBase) SetScrapeTime(scrapeTime time.Duration) {
	c.scrapeTime = &scrapeTime
}

func (c *CollectorBase) GetScrapeTime() *time.Duration {
	return c.scrapeTime
}

func (c *CollectorBase) SetAzureProjects(projects *devopsClient.ProjectList) {
	c.azureDevOpsProjectsMutex.Lock()
	c.azureDevOpsProjects = projects
	c.azureDevOpsProjectsMutex.Unlock()
}

func (c *CollectorBase) GetAzureProjects() (projects *devopsClient.ProjectList) {
	c.azureDevOpsProjectsMutex.Lock()
	projects = c.azureDevOpsProjects
	c.azureDevOpsProjectsMutex.Unlock()
	return
}

func (c *CollectorBase) collectionStart() {
	startTime := time.Now()
	c.collectionStartTime = &startTime

	if c.collectionLastTime == nil {
		lastTime := startTime.Add(-*c.GetScrapeTime())
		c.collectionLastTime = &lastTime
	}

	Logger.Infof("collector[%s]: starting metrics collection", c.Name)
}

func (c *CollectorBase) collectionFinish() {
	duration := time.Now().Sub(*c.collectionStartTime)
	c.LastScrapeDuration = &duration

	c.collectionLastTime = c.collectionStartTime

	Logger.Infof("collector[%s]: finished metrics collection (duration: %v)", c.Name, c.LastScrapeDuration)
}

func (c *CollectorBase) sleepUntilNextCollection() {
	Logger.Verbosef("collector[%s]: sleeping %v", c.Name, c.GetScrapeTime().String())
	time.Sleep(*c.GetScrapeTime())
}
