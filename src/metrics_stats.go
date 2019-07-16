package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MetricsCollectorStats struct {
	CollectorProcessorProject

	prometheus struct {
		agentPoolBuildCount      *prometheus.CounterVec
		agentPoolBuildWait       *prometheus.SummaryVec
		agentPoolBuildDuration   *prometheus.SummaryVec

		projectBuildCount        *prometheus.CounterVec
		projectBuildWait         *prometheus.SummaryVec
		projectBuildDuration     *prometheus.SummaryVec
	}

	session struct {
		agentPoolBuildWait *MetricCollectorList
		agentPoolBuildDuration *MetricCollectorList

		projectBuildWait *MetricCollectorList
		projectBuildDuration *MetricCollectorList
	}
}

func (m *MetricsCollectorStats) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	summaryMaxAgeSecs := (*m.CollectorReference.scrapeTime).Seconds() * 3
	summaryMaxAge, _ := time.ParseDuration(fmt.Sprintf("%vs", summaryMaxAgeSecs))

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

	m.prometheus.agentPoolBuildWait = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "azure_devops_stats_agentpool_builds_wait",
			Help: "Azure DevOps stats agentpool builds wait duration",
			MaxAge: summaryMaxAge,
		},
		[]string{
			"agentPoolID",
			"projectID",
			"result",
		},
	)

	m.prometheus.agentPoolBuildDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "azure_devops_stats_agentpool_builds_duration",
			Help: "Azure DevOps stats agentpool builds process duration",
			MaxAge: summaryMaxAge,
		},
		[]string{
			"agentPoolID",
			"projectID",
			"result",
		},
	)

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

	m.prometheus.projectBuildWait = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "azure_devops_stats_project_builds_wait",
			Help: "Azure DevOps stats project builds wait duration",
			MaxAge: summaryMaxAge,
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"result",
		},
	)

	m.prometheus.projectBuildDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "azure_devops_stats_project_builds_duration",
			Help: "Azure DevOps stats project builds process duration",
			MaxAge: summaryMaxAge,
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"result",
		},
	)

	prometheus.MustRegister(m.prometheus.agentPoolBuildCount)
	prometheus.MustRegister(m.prometheus.agentPoolBuildWait)
	prometheus.MustRegister(m.prometheus.agentPoolBuildDuration)

	prometheus.MustRegister(m.prometheus.projectBuildCount)
	prometheus.MustRegister(m.prometheus.projectBuildWait)
	prometheus.MustRegister(m.prometheus.projectBuildDuration)

	m.session.agentPoolBuildDuration = NewMetricCollectorList()
	m.session.projectBuildDuration = NewMetricCollectorList()
	m.session.agentPoolBuildWait = NewMetricCollectorList()
	m.session.projectBuildDuration = NewMetricCollectorList()
}

func (m *MetricsCollectorStats) Reset() {
}

func (m *MetricsCollectorStats) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	m.CollectBuilds(ctx, callback, project)
}

func (m *MetricsCollectorStats) CollectBuilds(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	minTime := *m.CollectorReference.collectionLastTime

	buildList, err := AzureDevopsClient.ListBuildHistoryWithStatus(project.Id, minTime, "completed")
	if err != nil {
		Logger.Errorf("project[%v]call[ListBuildHistory]: %v", project.Name, err)
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

		if build.FinishTime.Second() >= 0 {
			jobDuration := build.FinishTime.Sub(build.StartTime)

			m.session.agentPoolBuildDuration.AddDuration(prometheus.Labels{
				"agentPoolID": int64ToString(build.Queue.Pool.Id),
				"projectID":   build.Project.Id,
				"result":      build.Result,
			}, jobDuration)

			m.session.projectBuildDuration.AddDuration(prometheus.Labels{
				"projectID":         build.Project.Id,
				"buildDefinitionID": int64ToString(build.Definition.Id),
				"result":            build.Result,
			}, jobDuration)
		}

		if waitDuration >= 0 {

			m.session.agentPoolBuildWait.Add(prometheus.Labels{
				"agentPoolID": int64ToString(build.Queue.Pool.Id),
				"projectID":   build.Project.Id,
				"result":      build.Result,
			}, waitDuration)

			m.session.projectBuildDuration.Add(prometheus.Labels{
				"projectID":         build.Project.Id,
				"buildDefinitionID": int64ToString(build.Definition.Id),
				"result":            build.Result,
			}, waitDuration)
		}

		callback <- func() {
			m.session.agentPoolBuildDuration.SummarySet(m.prometheus.agentPoolBuildDuration)
			m.session.projectBuildDuration.SummarySet(m.prometheus.projectBuildDuration)
			m.session.agentPoolBuildWait.SummarySet(m.prometheus.agentPoolBuildWait)
			m.session.projectBuildDuration.SummarySet(m.prometheus.projectBuildDuration)
		}
	}
}
