package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"go.uber.org/zap"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorLatestBuild struct {
	collector.Processor

	prometheus struct {
		build       *prometheus.GaugeVec
		buildStatus *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorLatestBuild) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

	m.prometheus.build = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_latest_info",
			Help: "Azure DevOps build (latest)",
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"buildID",
			"agentPoolID",
			"requestedBy",
			"buildNumber",
			"buildName",
			"sourceBranch",
			"sourceVersion",
			"status",
			"reason",
			"result",
			"url",
		},
	)
	m.Collector.RegisterMetricList("build", m.prometheus.build, true)

	m.prometheus.buildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_latest_status",
			Help: "Azure DevOps build (latest)",
		},
		[]string{
			"projectID",
			"buildID",
			"buildNumber",
			"type",
		},
	)
	m.Collector.RegisterMetricList("buildStatus", m.prometheus.buildStatus, true)
}

func (m *MetricsCollectorLatestBuild) Reset() {}

func (m *MetricsCollectorLatestBuild) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))
		m.collectLatestBuilds(ctx, projectLogger, project, callback)
	}
}

func (m *MetricsCollectorLatestBuild) collectLatestBuilds(ctx context.Context, logger *zap.SugaredLogger, project devopsClient.Project, callback chan<- func()) {
	list, err := AzureDevopsClient.ListLatestBuilds(project.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	buildMetric := m.Collector.GetMetricList("build")
	buildStatusMetric := m.Collector.GetMetricList("buildStatus")

	for _, build := range list.List {
		buildMetric.AddInfo(prometheus.Labels{
			"projectID":         project.Id,
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildID":           int64ToString(build.Id),
			"buildNumber":       build.BuildNumber,
			"buildName":         build.Definition.Name,
			"agentPoolID":       int64ToString(build.Queue.Pool.Id),
			"requestedBy":       build.RequestedBy.DisplayName,
			"sourceBranch":      build.SourceBranch,
			"sourceVersion":     build.SourceVersion,
			"status":            build.Status,
			"reason":            build.Reason,
			"result":            build.Result,
			"url":               build.Links.Web.Href,
		})

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":   project.Id,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "started",
		}, build.StartTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":   project.Id,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "queued",
		}, build.QueueTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":   project.Id,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "finished",
		}, build.FinishTime)

		buildStatusMetric.AddDuration(prometheus.Labels{
			"projectID":   project.Id,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "jobDuration",
		}, build.FinishTime.Sub(build.StartTime))
	}
}
