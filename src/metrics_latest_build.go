package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorLatestBuild struct {
	CollectorProcessorProject

	prometheus struct {
		build       *prometheus.GaugeVec
		buildStatus *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorLatestBuild) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

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

	prometheus.MustRegister(m.prometheus.build)
	prometheus.MustRegister(m.prometheus.buildStatus)
}

func (m *MetricsCollectorLatestBuild) Reset() {
	m.prometheus.build.Reset()
	m.prometheus.buildStatus.Reset()
}

func (m *MetricsCollectorLatestBuild) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListLatestBuilds(project.Name)
	if err != nil {
		Logger.Errorf("project[%v]call[ListLatestBuilds]: %v", project.Name, err)
		return
	}

	buildMetric := MetricCollectorList{}
	buildStatusMetric := MetricCollectorList{}

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
	}

	callback <- func() {
		buildMetric.GaugeSet(m.prometheus.build)
		buildStatusMetric.GaugeSet(m.prometheus.buildStatus)
	}
}
