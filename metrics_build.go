package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	prometheusCommon "github.com/webdevops/go-prometheus-common"
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
	prometheus.MustRegister(m.prometheus.build)

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
	prometheus.MustRegister(m.prometheus.buildStatus)

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
	prometheus.MustRegister(m.prometheus.buildDefinition)
}

func (m *MetricsCollectorBuild) Reset() {
	m.prometheus.build.Reset()
	m.prometheus.buildDefinition.Reset()
	m.prometheus.buildStatus.Reset()
}

func (m *MetricsCollectorBuild) Collect(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	m.collectDefinition(ctx, logger, callback, project)
	m.collectBuilds(ctx, logger, callback, project)
}

func (m *MetricsCollectorBuild) collectDefinition(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListBuildDefinitions(project.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	buildDefinitonMetric := prometheusCommon.NewMetricsList()

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

func (m *MetricsCollectorBuild) collectBuilds(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	minTime := time.Now().Add(-opts.Limit.BuildHistoryDuration)

	list, err := AzureDevopsClient.ListBuildHistory(project.Id, minTime)
	if err != nil {
		logger.Error(err)
		return
	}

	buildMetric := prometheusCommon.NewMetricsList()
	buildStatusMetric := prometheusCommon.NewMetricsList()

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
