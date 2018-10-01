package main

import (
	"fmt"
	"time"
	"sync"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics run
func runMetricsCollectionGeneral() {
	var wg sync.WaitGroup

	callbackChannel := make(chan func())
	callbackAgentPools := make(chan devopsClient.AgentQueue)

	projectList, err := AzureDevopsClient.ListProjects()
	if err != nil {
		panic(err)
	}

	for _, project := range projectList.List {
		Logger.Messsage("project[%v]: starting metrics collection", project.Name)

		wg.Add(1)
		go func(project devopsClient.Project) {
			defer wg.Done()
			collectProject(project, callbackChannel)
			collectAgentQueues(project, callbackChannel, callbackAgentPools)
			collectBuilds(project, callbackChannel)
			collectBuildQueues(project, callbackChannel)
		}(project)

		repositoryList, err := AzureDevopsClient.ListRepositories(project.Name)
		if err != nil {
			ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		}

		for _, repository := range repositoryList.List {

			wg.Add(1)
			go func(project devopsClient.Project, repository devopsClient.Repository) {
				defer wg.Done()
				collectRepository(project, repository, callbackChannel)
				collectPullRequests(project, repository, callbackChannel)
			}(project, repository)
		}

		Logger.Messsage("project[%v]: found %v repositories", project.Name, repositoryList.Count)
	}

	go func() {
		var callbackList []func()
		for callback := range callbackChannel {
			callbackList = append(callbackList, callback)
		}

		prometheusProject.Reset()
		prometheusRepository.Reset()
		prometheusPullRequest.Reset()
		prometheusPullRequestStatus.Reset()
		prometheusAgentPool.Reset()
		prometheusAgentPoolSize.Reset()
		prometheusAgentPoolBuilds.Reset()
		prometheusAgentPoolWait.Reset()
		prometheusBuild.Reset()
		prometheusBuildStatus.Reset()
		for _, callback := range callbackList {
			callback()
		}

		Logger.Messsage("run[main]: finished")
	}()

	go func() {
		list := map[int64]devopsClient.AgentQueue{}
		for agentPool := range callbackAgentPools {
			list[agentPool.Id] = agentPool
		}

		agentPoolListMux.Lock()
		agentPoolList = list
		agentPoolListMux.Unlock()
	}()

	// wait for all funcs
	wg.Wait()
	close(callbackChannel)
	close(callbackAgentPools)
}


func collectProject(project devopsClient.Project, callback chan<- func()) {
	infoLabels := prometheus.Labels{
		"projectID": project.Id,
		"projectName": project.Name,
	}

	callback <- func() {
		prometheusProject.With(infoLabels).Set(1)
	}
}


func collectRepository(project devopsClient.Project, repository devopsClient.Repository, callback chan<- func()) {
	infoLabels := prometheus.Labels{
		"projectID": project.Id,
		"repositoryID": repository.Id,
		"repositoryName": repository.Name,
	}

	callback <- func() {
		prometheusRepository.With(infoLabels).Set(1)
	}
}


func collectBuilds(project devopsClient.Project, callback chan<- func()) {
	buildList, err := AzureDevopsClient.ListBuilds(project.Name)
	if err != nil {
		ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		return
	}

	for _, build := range buildList.List {
		infoLabels := prometheus.Labels{
			"projectID": project.Id,
			"buildID": fmt.Sprintf("%d", build.Id),
			"buildNumber": build.BuildNumber,
			"buildName": build.Definition.Name,
			"agentPoolID": fmt.Sprintf("%d", build.Queue.Pool.Id),
			"requestedBy": build.RequestedBy.DisplayName,
			"status": build.Status,
			"reason": build.Reason,
			"result": build.Result,
			"url": build.Links.Web.Href,
		}

		statusStartedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"buildID": fmt.Sprintf("%d", build.Id),
			"buildNumber": build.BuildNumber,
			"type": "started",
		}
		statusStartedValue := float64(build.StartTime.Unix())

		statuQueuedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"buildID": fmt.Sprintf("%d", build.Id),
			"buildNumber": build.BuildNumber,
			"type": "queued",
		}
		statusQueuedValue := float64(build.QueueTime.Unix())

		statuFinishedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"buildID": fmt.Sprintf("%d", build.Id),
			"buildNumber": build.BuildNumber,
			"type": "finished",
		}
		statusFinishedValue := float64(build.FinishTime.Unix())

		callback <- func() {
			prometheusBuild.With(infoLabels).Set(1)
			prometheusBuildStatus.With(statuQueuedLabels).Set(statusQueuedValue)
			prometheusBuildStatus.With(statusStartedLabels).Set(statusStartedValue)
			prometheusBuildStatus.With(statuFinishedLabels).Set(statusFinishedValue)
		}
	}
}

func collectBuildQueues(project devopsClient.Project, callback chan<- func()) {
	minTime := time.Now().Add(- opts.ScrapeTime)

	buildList, err := AzureDevopsClient.ListBuildHistory(project.Name, minTime)
	if err != nil {
		ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		return
	}

	for _, build := range buildList.List {
		waitDuration := build.QueueDuration().Seconds()

		agentPoolBuildLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", build.Queue.Pool.Id),
			"result": build.Result,
		}

		agentPoolWaitLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", build.Queue.Pool.Id),
		}

		callback <- func() {
			prometheusAgentPoolBuilds.With(agentPoolBuildLabels).Inc()
			prometheusAgentPoolWait.With(agentPoolWaitLabels).Observe(waitDuration)
		}
	}
}

func collectAgentQueues(project devopsClient.Project, callback chan<- func(), callbackAgentPools chan<- devopsClient.AgentQueue) {
	agentQueueList, err := AzureDevopsClient.ListAgentQueues(project.Name)
	if err != nil {
		ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		return
	}


	for _, agentQueue := range agentQueueList.List {

		infoLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", agentQueue.Pool.Id),
			"agentPoolName": agentQueue.Name,
			"isHosted": boolToString(agentQueue.Pool.IsHosted),
			"agentPoolType": agentQueue.Pool.PoolType,
		}

		sizeLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", agentQueue.Pool.Id),
		}
		agentPoolValue := float64(agentQueue.Pool.Size)

		callback <- func() {
			prometheusAgentPool.With(infoLabels).Set(1)
			prometheusAgentPoolSize.With(sizeLabels).Set(agentPoolValue)
		}

		callbackAgentPools <- agentQueue
	}
}

func collectPullRequests(project devopsClient.Project, repository devopsClient.Repository, callback chan<- func()) {
	list, err := AzureDevopsClient.ListPullrequest(project.Name, repository.Id)
	if err != nil {
		ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		return
	}

	for _, pullRequest := range list.List {
		infoLabels := prometheus.Labels{
			"projectID":     project.Id,
			"repositoryID":  repository.Id,
			"pullrequestID": fmt.Sprintf("%d", pullRequest.Id),
			"pullrequestTitle": pullRequest.Title,
			"status": pullRequest.Status,
			"creator": pullRequest.CreatedBy.DisplayName,
			"sourceBranch": pullRequest.SourceRefName,
			"targetBranch": pullRequest.TargetRefName,
		}

		statusCreatedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"repositoryID":  repository.Id,
			"pullrequestID": fmt.Sprintf("%d", pullRequest.Id),
			"type": "created",
		}
		statusCreatedValue := float64(pullRequest.CreationDate.Unix())

		callback <- func() {
			prometheusPullRequest.With(infoLabels).Set(1)
			prometheusPullRequestStatus.With(statusCreatedLabels).Set(statusCreatedValue)
		}
	}
}
