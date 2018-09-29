package main

import (
	"fmt"
	"time"
	"sync"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	prometheusProject *prometheus.GaugeVec
	prometheusRepository *prometheus.GaugeVec
	prometheusPullRequest *prometheus.GaugeVec
	prometheusPullRequestStatus *prometheus.GaugeVec
	prometheusAgentPool *prometheus.GaugeVec
	prometheusAgentPoolBuilds *prometheus.GaugeVec
	prometheusAgentPoolWait *prometheus.SummaryVec
	prometheusBuild *prometheus.GaugeVec
	prometheusBuildStatus *prometheus.GaugeVec
)

// Create and setup metrics and collection
func setupMetricsCollection() {
	prometheusProject = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_project_info",
			Help: "Azure DevOps project",
		},
		[]string{"projectID", "projectName"},
	)

	prometheusRepository = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_repository_info",
			Help: "Azure DevOps repository",
		},
		[]string{"projectID", "repositoryID", "repositoryName"},
	)

	prometheusPullRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_pullrequest_info",
			Help: "Azure DevOps pullrequest",
		},
		[]string{"projectID", "repositoryID", "pullrequestID", "pullrequestTitle", "sourceBranch", "targetBranch", "status", "creator"},
	)

	prometheusPullRequestStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_pullrequest_status",
			Help: "Azure DevOps pullrequest",
		},
		[]string{"projectID", "repositoryID", "pullrequestID", "type"},
	)

	prometheusAgentPool = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_info",
			Help: "Azure DevOps agentpool",
		},
		[]string{"agentPoolID", "agentPoolName"},
	)

	prometheusAgentPoolBuilds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_builds",
			Help: "Azure DevOps agentpool",
		},
		[]string{"agentPoolID", "result"},
	)

	prometheusAgentPoolWait = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "azure_devops_agentpool_wait",
			Help: "Azure DevOps agentpool waittime",
		},
		[]string{"agentPoolID"},
	)

	prometheusBuild = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_info",
			Help: "Azure DevOps build",
		},
		[]string{"projectID", "buildID", "agentPoolID", "requestedBy", "buildNumber", "buildName", "status", "reason", "result", "url"},
	)

	prometheusBuildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_status",
			Help: "Azure DevOps build",
		},
		[]string{"projectID", "buildID", "buildNumber", "type"},
	)

	prometheus.MustRegister(prometheusProject)
	prometheus.MustRegister(prometheusRepository)
	prometheus.MustRegister(prometheusPullRequest)
	prometheus.MustRegister(prometheusPullRequestStatus)
	prometheus.MustRegister(prometheusAgentPool)
	prometheus.MustRegister(prometheusAgentPoolBuilds)
	prometheus.MustRegister(prometheusAgentPoolWait)
	prometheus.MustRegister(prometheusBuild)
	prometheus.MustRegister(prometheusBuildStatus)
}

// Start backgrounded metrics collection
func startMetricsCollection() {
	go func() {
		for {
			go func() {
				runMetricsCollection()
			}()
			time.Sleep(opts.ScrapeTime)
		}
	}()
}

// Metrics run
func runMetricsCollection() {
	var wg sync.WaitGroup

	callbackChannel := make(chan func())


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
		prometheusAgentPoolBuilds.Reset()
		prometheusAgentPoolWait.Reset()
		prometheusBuild.Reset()
		prometheusBuildStatus.Reset()
		for _, callback := range callbackList {
			callback()
		}

		Logger.Messsage("run: finished")
	}()

	// wait for all funcs
	wg.Wait()
	close(callbackChannel)
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

		agentPoolLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", build.Queue.Pool.Id),
			"agentPoolName": build.Queue.Name,
		}

		callback <- func() {
			prometheusBuild.With(infoLabels).Set(1)
			prometheusBuildStatus.With(statuQueuedLabels).Set(statusQueuedValue)
			prometheusBuildStatus.With(statusStartedLabels).Set(statusStartedValue)
			prometheusBuildStatus.With(statuFinishedLabels).Set(statusFinishedValue)

			prometheusAgentPool.With(agentPoolLabels).Set(1)
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

		agentPoolLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", build.Queue.Pool.Id),
			"agentPoolName": build.Queue.Name,
		}

		agentPoolBuildLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", build.Queue.Pool.Id),
			"result": build.Result,
		}

		agentPoolWaitLabels := prometheus.Labels{
			"agentPoolID": fmt.Sprintf("%d", build.Queue.Pool.Id),
		}

		callback <- func() {
			prometheusAgentPool.With(agentPoolLabels).Set(1)
			prometheusAgentPoolBuilds.With(agentPoolBuildLabels).Inc()
			prometheusAgentPoolWait.With(agentPoolWaitLabels).Observe(waitDuration)
		}
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
