package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"go.uber.org/zap"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorStats struct {
	collector.Processor

	prometheus struct {
		agentPoolBuildCount    *prometheus.CounterVec
		agentPoolBuildWait     *prometheus.SummaryVec
		agentPoolBuildDuration *prometheus.SummaryVec

		projectBuildCount      *prometheus.CounterVec
		projectBuildWait       *prometheus.SummaryVec
		projectBuildDuration   *prometheus.SummaryVec
		projectBuildSuccess    *prometheus.SummaryVec
		projectReleaseDuration *prometheus.SummaryVec
		projectReleaseSuccess  *prometheus.SummaryVec
	}
}

func (m *MetricsCollectorStats) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

	// ------------------------------------------
	// AgentPool
	// ------------------------------------------

	m.prometheus.agentPoolBuildCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "azure_devops_stats_agentpool_builds",
			Help: "Azure DevOps stats agentpool builds counter",
		},
		[]string{
			"agentPoolID",
			"projectID",
			"result",
		},
	)
	m.Collector.RegisterMetricList("agentPoolBuildCount", m.prometheus.agentPoolBuildCount, false)

	m.prometheus.agentPoolBuildWait = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "azure_devops_stats_agentpool_builds_wait",
			Help:   "Azure DevOps stats agentpool builds wait duration",
			MaxAge: *opts.Stats.SummaryMaxAge,
		},
		[]string{
			"agentPoolID",
			"projectID",
			"result",
		},
	)
	m.Collector.RegisterMetricList("agentPoolBuildWait", m.prometheus.agentPoolBuildWait, false)

	m.prometheus.agentPoolBuildDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "azure_devops_stats_agentpool_builds_duration",
			Help:   "Azure DevOps stats agentpool builds process duration",
			MaxAge: *opts.Stats.SummaryMaxAge,
		},
		[]string{
			"agentPoolID",
			"projectID",
			"result",
		},
	)
	m.Collector.RegisterMetricList("agentPoolBuildDuration", m.prometheus.agentPoolBuildDuration, false)

	// ------------------------------------------
	// Project
	// ------------------------------------------

	m.prometheus.projectBuildCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "azure_devops_stats_project_builds",
			Help: "Azure DevOps stats project builds counter",
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"result",
		},
	)
	m.Collector.RegisterMetricList("projectBuildCount", m.prometheus.projectBuildCount, false)

	m.prometheus.projectBuildSuccess = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "azure_devops_stats_project_success",
			Help: "Azure DevOps stats project success",
		},
		[]string{
			"projectID",
			"buildDefinitionID",
		},
	)
	m.Collector.RegisterMetricList("projectBuildSuccess", m.prometheus.projectBuildSuccess, false)

	m.prometheus.projectBuildWait = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "azure_devops_stats_project_builds_wait",
			Help:   "Azure DevOps stats project builds wait duration",
			MaxAge: *opts.Stats.SummaryMaxAge,
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"result",
		},
	)
	m.Collector.RegisterMetricList("projectBuildWait", m.prometheus.projectBuildWait, false)

	m.prometheus.projectBuildDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "azure_devops_stats_project_builds_duration",
			Help:   "Azure DevOps stats project builds process duration",
			MaxAge: *opts.Stats.SummaryMaxAge,
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"result",
		},
	)
	m.Collector.RegisterMetricList("projectBuildDuration", m.prometheus.projectBuildDuration, false)

	m.prometheus.projectReleaseDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "azure_devops_stats_project_release_duration",
			Help:   "Azure DevOps stats project release process duration",
			MaxAge: *opts.Stats.SummaryMaxAge,
		},
		[]string{
			"projectID",
			"releaseDefinitionID",
			"definitionEnvironmentID",
			"status",
		},
	)
	m.Collector.RegisterMetricList("projectReleaseDuration", m.prometheus.projectReleaseDuration, false)

	m.prometheus.projectReleaseSuccess = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "azure_devops_stats_project_release_success",
			Help:   "Azure DevOps stats project release success",
			MaxAge: *opts.Stats.SummaryMaxAge,
		},
		[]string{
			"projectID",
			"releaseDefinitionID",
			"definitionEnvironmentID",
		},
	)
	m.Collector.RegisterMetricList("projectReleaseSuccess", m.prometheus.projectReleaseSuccess, false)
}

func (m *MetricsCollectorStats) Reset() {}

func (m *MetricsCollectorStats) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))
		m.CollectBuilds(ctx, projectLogger, callback, project)
		m.CollectReleases(ctx, projectLogger, callback, project)
	}
}

func (m *MetricsCollectorStats) CollectReleases(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), project devopsClient.Project) {
	minTime := time.Now().Add(-*m.Collector.GetScapeTime())
	if val := m.Collector.GetLastScapeTime(); val != nil {
		minTime = *val
	}

	releaseList, err := AzureDevopsClient.ListReleaseHistory(project.Id, minTime)
	if err != nil {
		logger.Error(err)
		return
	}

	for _, release := range releaseList.List {
		for _, environment := range release.Environments {
			switch environment.Status {
			case "succeeded":
				m.prometheus.projectReleaseSuccess.With(prometheus.Labels{
					"projectID":               release.Project.Id,
					"releaseDefinitionID":     int64ToString(release.Definition.Id),
					"definitionEnvironmentID": int64ToString(environment.DefinitionEnvironmentId),
				}).Observe(1)
			case "failed", "partiallySucceeded":
				m.prometheus.projectReleaseSuccess.With(prometheus.Labels{
					"projectID":               release.Project.Id,
					"releaseDefinitionID":     int64ToString(release.Definition.Id),
					"definitionEnvironmentID": int64ToString(environment.DefinitionEnvironmentId),
				}).Observe(0)
			}

			timeToDeploy := environment.TimeToDeploy * 60
			if timeToDeploy > 0 {
				m.prometheus.projectReleaseDuration.With(prometheus.Labels{
					"projectID":               release.Project.Id,
					"releaseDefinitionID":     int64ToString(release.Definition.Id),
					"definitionEnvironmentID": int64ToString(environment.DefinitionEnvironmentId),
					"status":                  environment.Status,
				}).Observe(timeToDeploy)
			}
		}
	}
}

func (m *MetricsCollectorStats) CollectBuilds(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), project devopsClient.Project) {
	minTime := time.Now().Add(-opts.Limit.BuildHistoryDuration)

	buildList, err := AzureDevopsClient.ListBuildHistoryWithStatus(project.Id, minTime, "completed")
	if err != nil {
		logger.Error(err)
		return
	}

	for _, build := range buildList.List {
		waitDuration := build.QueueDuration().Seconds()

		m.prometheus.agentPoolBuildCount.With(prometheus.Labels{
			"agentPoolID": int64ToString(build.Queue.Pool.Id),
			"projectID":   build.Project.Id,
			"result":      build.Result,
		}).Inc()

		m.prometheus.projectBuildCount.With(prometheus.Labels{
			"projectID":         build.Project.Id,
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"result":            build.Result,
		}).Inc()

		switch build.Result {
		case "succeeded":
			m.prometheus.projectBuildSuccess.With(prometheus.Labels{
				"projectID":         build.Project.Id,
				"buildDefinitionID": int64ToString(build.Definition.Id),
			}).Observe(1)
		case "failed":
			m.prometheus.projectBuildSuccess.With(prometheus.Labels{
				"projectID":         build.Project.Id,
				"buildDefinitionID": int64ToString(build.Definition.Id),
			}).Observe(0)
		}

		if build.FinishTime.Second() >= 0 {
			jobDuration := build.FinishTime.Sub(build.StartTime)

			m.prometheus.agentPoolBuildDuration.With(prometheus.Labels{
				"agentPoolID": int64ToString(build.Queue.Pool.Id),
				"projectID":   build.Project.Id,
				"result":      build.Result,
			}).Observe(jobDuration.Seconds())

			m.prometheus.projectBuildDuration.With(prometheus.Labels{
				"projectID":         build.Project.Id,
				"buildDefinitionID": int64ToString(build.Definition.Id),
				"result":            build.Result,
			}).Observe(jobDuration.Seconds())
		}

		if waitDuration >= 0 {
			m.prometheus.agentPoolBuildWait.With(prometheus.Labels{
				"agentPoolID": int64ToString(build.Queue.Pool.Id),
				"projectID":   build.Project.Id,
				"result":      build.Result,
			}).Observe(waitDuration)

			m.prometheus.projectBuildWait.With(prometheus.Labels{
				"projectID":         build.Project.Id,
				"buildDefinitionID": int64ToString(build.Definition.Id),
				"result":            build.Result,
			}).Observe(waitDuration)
		}
	}
}
