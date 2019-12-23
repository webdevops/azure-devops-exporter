package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	"time"
)

type MetricsCollectorBuild struct {
	CollectorProcessorProject

	prometheus struct {
		build       *prometheus.GaugeVec
		buildStatus *prometheus.GaugeVec

		buildDefinition *prometheus.GaugeVec

		buildTimeProject *prometheus.SummaryVec
		jobTimeProject   *prometheus.SummaryVec
	}
}

func (m *MetricsCollectorBuild) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.build = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_info",
			Help: "Azure DevOps build",
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

	m.prometheus.buildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_status",
			Help: "Azure DevOps build",
		},
		[]string{
			"projectID",
			"buildID",
			"buildDefinitionID",
			"buildNumber",
			"type",
		},
	)

	m.prometheus.buildDefinition = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_definition_info",
			Help: "Azure DevOps build definition",
		},
		[]string{
			"projectID",
			"buildDefinitionID",
			"buildNameFormat",
			"buildDefinitionName",
			"path",
			"url",
		},
	)

	prometheus.MustRegister(m.prometheus.build)
	prometheus.MustRegister(m.prometheus.buildStatus)
	prometheus.MustRegister(m.prometheus.buildDefinition)
}

func (m *MetricsCollectorBuild) Reset() {
	m.prometheus.build.Reset()
	m.prometheus.buildDefinition.Reset()
	m.prometheus.buildStatus.Reset()
}

func (m *MetricsCollectorBuild) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	m.collectDefinition(ctx, callback, project)
	m.collectBuilds(ctx, callback, project)

}

func (m *MetricsCollectorBuild) collectDefinition(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListBuildDefinitions(project.Id)
	if err != nil {
		Logger.Errorf("project[%v]call[ListBuildDefinitions]: %v", project.Name, err)
		return
	}

	buildDefinitonMetric := NewMetricCollectorList()

	for _, buildDefinition := range list.List {
		buildDefinitonMetric.Add(prometheus.Labels{
			"projectID":           project.Id,
			"buildDefinitionID":   int64ToString(buildDefinition.Id),
			"buildNameFormat":     buildDefinition.BuildNameFormat,
			"buildDefinitionName": buildDefinition.Name,
			"path":                buildDefinition.Path,
			"url":                 buildDefinition.Links.Web.Href,
		}, 1)
	}

	callback <- func() {
		buildDefinitonMetric.GaugeSet(m.prometheus.buildDefinition)
	}
}

func (m *MetricsCollectorBuild) collectBuilds(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	minTime := time.Now().Add(-opts.LimitBuildHistoryDuration)

	list, err := AzureDevopsClient.ListBuildHistory(project.Id, minTime)
	if err != nil {
		Logger.Errorf("project[%v]call[ListBuildHistory]: %v", project.Name, err)
		return
	}

	buildMetric := NewMetricCollectorList()
	buildStatusMetric := NewMetricCollectorList()

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

		buildStatusMetric.AddBool(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"type":              "succeeded",
		}, build.Result == "succeeded")

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"type":              "queued",
		}, build.QueueTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"type":              "started",
		}, build.StartTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"type":              "finished",
		}, build.FinishTime)

		buildStatusMetric.AddDuration(prometheus.Labels{
			"projectID":         project.Id,
			"buildID":           int64ToString(build.Id),
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildNumber":       build.BuildNumber,
			"type":              "jobDuration",
		}, build.FinishTime.Sub(build.StartTime))
	}

	callback <- func() {
		buildMetric.GaugeSet(m.prometheus.build)
		buildStatusMetric.GaugeSet(m.prometheus.buildStatus)
	}
}
