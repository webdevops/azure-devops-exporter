package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorLatestBuild struct {
	CollectorProcessorProject

	prometheus struct {
		build *prometheus.GaugeVec
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
		[]string{"projectID", "buildDefinitionID", "buildID", "agentPoolID", "requestedBy", "buildNumber", "buildName", "sourceBranch", "sourceVersion", "status", "reason", "result", "url"},
	)

	m.prometheus.buildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_latest_status",
			Help: "Azure DevOps build (latest)",
		},
		[]string{"projectID", "buildID", "buildNumber", "type"},
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
