package main

import (
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	AzureDevops "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

const (
	azureDevopsServiceDiscoveryCacheKeyProjectList   = "projects"
	azureDevopsServiceDiscoveryCacheKeyAgentPoolList = "agentpools"
)

type (
	azureDevopsServiceDiscovery struct {
		cache       *cache.Cache
		cacheExpiry time.Duration

		logger *zap.SugaredLogger

		lock struct {
			projectList   sync.Mutex
			agentpoolList sync.Mutex
		}
	}
)

func NewAzureDevopsServiceDiscovery() *azureDevopsServiceDiscovery {
	sd := &azureDevopsServiceDiscovery{}
	sd.cacheExpiry = opts.ServiceDiscovery.RefreshDuration
	sd.cache = cache.New(sd.cacheExpiry, time.Duration(1*time.Minute))
	sd.logger = logger.With(zap.String("component", "servicediscovery"))

	sd.logger.Infof("init AzureDevops servicediscovery with %v cache", sd.cacheExpiry.String())
	return sd
}

func (sd *azureDevopsServiceDiscovery) Update() {
	sd.cache.Flush()
	sd.ProjectList()
	sd.AgentPoolList()
}

func (sd *azureDevopsServiceDiscovery) ProjectList() (list []AzureDevops.Project) {
	sd.lock.projectList.Lock()
	defer sd.lock.projectList.Unlock()

	if val, ok := sd.cache.Get(azureDevopsServiceDiscoveryCacheKeyProjectList); ok {
		// fetched from cache
		list = val.([]AzureDevops.Project)
		return
	}

	// cache was invalid, fetch data from api
	sd.logger.Infof("updating project list")
	result, err := AzureDevopsClient.ListProjects()
	if err != nil {
		sd.logger.Panic(err)
	}

	sd.logger.Infof("fetched %v projects", result.Count)

	list = result.List

	// whitelist
	if len(opts.AzureDevops.FilterProjects) > 0 {
		rawList := list
		list = []AzureDevops.Project{}
		for _, project := range rawList {
			if arrayStringContains(opts.AzureDevops.FilterProjects, project.Id) {
				list = append(list, project)
			}
		}
	}

	// blacklist
	if len(opts.AzureDevops.BlacklistProjects) > 0 {
		// filter ignored azure devops projects
		rawList := list
		list = []AzureDevops.Project{}
		for _, project := range rawList {
			if !arrayStringContains(opts.AzureDevops.BlacklistProjects, project.Id) {
				list = append(list, project)
			}
		}
	}

	// save to cache
	sd.cache.SetDefault(azureDevopsServiceDiscoveryCacheKeyProjectList, list)

	return
}

func (sd *azureDevopsServiceDiscovery) AgentPoolList() (list []int64) {
	sd.lock.agentpoolList.Lock()
	defer sd.lock.agentpoolList.Unlock()

	if val, ok := sd.cache.Get(azureDevopsServiceDiscoveryCacheKeyAgentPoolList); ok {
		// fetched from cache
		list = val.([]int64)
		return
	}

	if opts.AzureDevops.AgentPoolIdList != nil {
		sd.logger.Infof("using predefined AgentPool list")
		list = *opts.AzureDevops.AgentPoolIdList
	} else {
		sd.logger.Infof("upading AgentPool list")

		result, err := AzureDevopsClient.ListAgentPools()
		if err != nil {
			sd.logger.Panic(err)
			return
		}
		sd.logger.Infof("fetched %v agentpools", result.Count)

		for _, agentPool := range result.Value {
			list = append(list, agentPool.ID)
		}
	}

	// save to cache
	sd.cache.SetDefault(azureDevopsServiceDiscoveryCacheKeyAgentPoolList, list)

	return
}
