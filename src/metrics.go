package main

import (
	"sync"
	"time"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	prometheusProject *prometheus.GaugeVec
	prometheusRepository *prometheus.GaugeVec
	prometheusPullRequest *prometheus.GaugeVec
	prometheusPullRequestStatus *prometheus.GaugeVec
	prometheusAgentPool *prometheus.GaugeVec
	prometheusAgentPoolSize *prometheus.GaugeVec
	prometheusAgentPoolBuilds *prometheus.GaugeVec
	prometheusAgentPoolWait *prometheus.SummaryVec
	prometheusAgentPoolAgent *prometheus.GaugeVec
	prometheusAgentPoolAgentStatus *prometheus.GaugeVec
	prometheusAgentPoolAgentJob *prometheus.GaugeVec
	prometheusBuild *prometheus.GaugeVec
	prometheusBuildStatus *prometheus.GaugeVec

	agentPoolList map[int64]devopsClient.AgentQueue
	agentPoolListMux sync.Mutex
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
		[]string{"agentPoolID", "agentPoolName", "agentPoolType", "isHosted"},
	)

	prometheusAgentPoolSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_size",
			Help: "Azure DevOps agentpool",
		},
		[]string{"agentPoolID"},
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

	prometheusAgentPoolAgent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_agent_info",
			Help: "Azure DevOps agentpool",
		},
		[]string{"agentPoolID", "agentPoolAgentID", "agentPoolAgentName", "agentPoolAgentVersion", "provisioningState", "maxParallelism", "agentPoolAgentOs"},
	)

	prometheusAgentPoolAgentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_agent_status",
			Help: "Azure DevOps agentpool",
		},
		[]string{"agentPoolAgentID", "type"},
	)

	prometheusAgentPoolAgentJob = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_agent_job",
			Help: "Azure DevOps agentpool",
		},
		[]string{"agentPoolAgentID", "jobRequestId"},
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
	prometheus.MustRegister(prometheusAgentPoolSize)
	prometheus.MustRegister(prometheusAgentPoolBuilds)
	prometheus.MustRegister(prometheusAgentPoolWait)
	prometheus.MustRegister(prometheusAgentPoolAgent)
	prometheus.MustRegister(prometheusAgentPoolAgentStatus)
	prometheus.MustRegister(prometheusAgentPoolAgentJob)
	prometheus.MustRegister(prometheusBuild)
	prometheus.MustRegister(prometheusBuildStatus)
}

// Start backgrounded metrics collection
func startMetricsCollection() {
	go func() {
		for {
			go func() {
				runMetricsCollectionGeneral()
			}()
			time.Sleep(opts.ScrapeTime)
		}
	}()

	go func() {
		for {
			go func() {
				runMetricsCollectionAgentPool()
			}()
			time.Sleep(opts.ScrapeTimeQueues)
		}
	}()
}
