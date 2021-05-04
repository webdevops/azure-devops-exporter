package main

import (
	log "github.com/sirupsen/logrus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	"sync"
	"time"
)

type CollectorBase struct {
	Name       string
	scrapeTime *time.Duration

	azureDevOpsProjects      *devopsClient.ProjectList
	azureDevOpsProjectsMutex sync.Mutex

	logger *log.Entry

	LastScrapeDuration  *time.Duration
	collectionStartTime *time.Time
	collectionLastTime  *time.Time
}

func (c *CollectorBase) Init() {
	c.logger = log.WithField("collector", c.Name)

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

	c.logger.Info("starting metrics collection")
}

func (c *CollectorBase) collectionFinish() {
	duration := time.Since(*c.collectionStartTime)
	c.LastScrapeDuration = &duration

	c.collectionLastTime = c.collectionStartTime

	c.logger.WithField("duration", c.LastScrapeDuration.Seconds()).Infof("finished metrics collection (duration: %v)", c.LastScrapeDuration)
}

func (c *CollectorBase) sleepUntilNextCollection() {
	c.logger.Debugf("sleeping %v", c.GetScrapeTime().String())
	time.Sleep(*c.GetScrapeTime())
}
