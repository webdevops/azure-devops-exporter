package main

import (
	"sync"
	"time"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
)

type CollectorBase struct {
	Name string
	scrapeTime  *time.Duration

	azureDevOpsProjects *devopsClient.ProjectList
	azureDevOpsProjectsMutex sync.Mutex
}

func (c *CollectorBase) Init() {
}

func (c *CollectorBase) SetScrapeTime(scrapeTime time.Duration) {
	c.scrapeTime = &scrapeTime
}

func (c *CollectorBase) GetScrapeTime() (*time.Duration) {
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
