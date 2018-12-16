package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorBuild struct {
	CollectorProcessorProject

	prometheus struct {
		build *prometheus.GaugeVec
		buildDefinition *prometheus.GaugeVec
		buildStatus *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorBuild) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.build = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_info",
			Help: "Azure DevOps build",
		},
		[]string{"projectID", "buildDefinitionID", "buildID", "agentPoolID", "requestedBy", "buildNumber", "buildName", "sourceBranch", "sourceVersion", "status", "reason", "result", "url"},
	)

	m.prometheus.buildDefinition = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_definition",
			Help: "Azure DevOps build definition",
		},
		[]string{"projectID", "buildDefinitionID", "buildNameFormat", "buildDefinitionName", "url"},
	)

	m.prometheus.buildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_status",
			Help: "Azure DevOps build",
		},
		[]string{"projectID", "buildID", "buildNumber", "type"},
	)

	prometheus.MustRegister(m.prometheus.build)
	prometheus.MustRegister(m.prometheus.buildDefinition)
	prometheus.MustRegister(m.prometheus.buildStatus)
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
	list, err := AzureDevopsClient.ListBuildDefinitions(project.Name)
	if err != nil {
		ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		return
	}

	buildDefinitonMetric := MetricCollectorList{}

	for _, buildDefinition := range list.List {
		infoLabels := prometheus.Labels{
			"projectID":           project.Id,
			"buildDefinitionID":   int64ToString(buildDefinition.Id),
			"buildNameFormat":     buildDefinition.BuildNameFormat,
			"buildDefinitionName": buildDefinition.Name,
			"url":                 buildDefinition.Links.Web.Href,
		}
		buildDefinitonMetric.Add(infoLabels, 1)
	}

	callback <- func() {
		buildDefinitonMetric.GaugeSet(m.prometheus.buildDefinition)
	}
}

func (m *MetricsCollectorBuild) collectBuilds(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListBuilds(project.Name)
	if err != nil {
		ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		return
	}

	buildMetric := MetricCollectorList{}
	buildStatusMetric := MetricCollectorList{}

	for _, build := range list.List {
		infoLabels := prometheus.Labels{
			"projectID": project.Id,
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildID": int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"buildName": build.Definition.Name,
			"agentPoolID": int64ToString(build.Queue.Pool.Id),
			"requestedBy": build.RequestedBy.DisplayName,
			"sourceBranch": build.SourceBranch,
			"sourceVersion": build.SourceVersion,
			"status": build.Status,
			"reason": build.Reason,
			"result": build.Result,
			"url": build.Links.Web.Href,
		}
		buildMetric.Add(infoLabels, 1)

		statusStartedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"buildID": int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type": "started",
		}
		statusStartedValue := float64(build.StartTime.Unix())
		if statusStartedValue > 0 {
			buildStatusMetric.Add(statusStartedLabels, statusStartedValue)
		}

		statusQueuedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"buildID": int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type": "queued",
		}
		statusQueuedValue := float64(build.QueueTime.Unix())
		if statusQueuedValue > 0 {
			buildStatusMetric.Add(statusQueuedLabels, statusQueuedValue)
		}

		statuFinishedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"buildID": int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type": "finished",
		}
		statusFinishedValue := float64(build.FinishTime.Unix())
		if statusFinishedValue > 0 {
			buildStatusMetric.Add(statuFinishedLabels, statusFinishedValue)
		}
	}

	callback <- func() {
		buildMetric.GaugeSet(m.prometheus.build)
		buildStatusMetric.GaugeSet(m.prometheus.buildStatus)
	}
}
