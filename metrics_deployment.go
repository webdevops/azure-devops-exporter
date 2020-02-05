package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorDeployment struct {
	CollectorProcessorProject

	prometheus struct {
		deployment       *prometheus.GaugeVec
		deploymentStatus *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorDeployment) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.deployment = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_deployment_info",
			Help: "Azure DevOps deployment",
		},
		[]string{
			"projectID",
			"deploymentID",
			"releaseID",
			"releaseName",
			"releaseDefinitionID",
			"requestedBy",
			"deploymentName",
			"deploymentStatus",
			"operationStatus",
			"reason",
			"attempt",
			"environmentId",
			"environmentName",
			"approvedBy",
		},
	)

	m.prometheus.deploymentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_deployment_status",
			Help: "Azure DevOps deployment status",
		},
		[]string{
			"projectID",
			"deploymentID",
			"type",
		},
	)

	prometheus.MustRegister(m.prometheus.deployment)
	prometheus.MustRegister(m.prometheus.deploymentStatus)
}

func (m *MetricsCollectorDeployment) Reset() {
	m.prometheus.deployment.Reset()
	m.prometheus.deploymentStatus.Reset()
}

func (m *MetricsCollectorDeployment) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListReleaseDefinitions(project.Id)
	if err != nil {
		Logger.Errorf("project[%v]call[ListReleaseDefinitions]: %v", project.Name, err)
		return
	}

	deploymentMetric := NewMetricCollectorList()
	deploymentStatusMetric := NewMetricCollectorList()

	for _, releaseDefinition := range list.List {
		deploymentList, err := AzureDevopsClient.ListReleaseDeployments(project.Id, releaseDefinition.Id)
		if err != nil {
			Logger.Errorf("project[%v]call[ListReleaseDeployments]: %v", project.Name, err)
			return
		}

		for _, deployment := range deploymentList.List {
			deploymentMetric.AddInfo(prometheus.Labels{
				"projectID":           project.Id,
				"deploymentID":        int64ToString(deployment.Id),
				"releaseID":           int64ToString(deployment.Release.Id),
				"releaseName":         deployment.Release.Name,
				"releaseDefinitionID": int64ToString(releaseDefinition.Id),
				"requestedBy":         deployment.RequestedBy.DisplayName,
				"deploymentName":      deployment.Name,
				"deploymentStatus":    deployment.DeploymentStatus,
				"operationStatus":     deployment.OperationStatus,
				"reason":              deployment.Reason,
				"attempt":             int64ToString(deployment.Attempt),
				"environmentId":       int64ToString(deployment.ReleaseEnvironment.Id),
				"environmentName":     deployment.ReleaseEnvironment.Name,
				"approvedBy":          deployment.ApprovedBy(),
			})

			queuedOn := deployment.QueuedOnTime()
			startedOn := deployment.StartedOnTime()
			completedOn := deployment.CompletedOnTime()

			if queuedOn != nil {
				deploymentStatusMetric.AddTime(prometheus.Labels{
					"projectID":    project.Id,
					"deploymentID": int64ToString(deployment.Id),
					"type":         "queued",
				}, *queuedOn)
			}

			if startedOn != nil {
				deploymentStatusMetric.AddTime(prometheus.Labels{
					"projectID":    project.Id,
					"deploymentID": int64ToString(deployment.Id),
					"type":         "started",
				}, *startedOn)
			}

			if completedOn != nil {
				deploymentStatusMetric.AddTime(prometheus.Labels{
					"projectID":    project.Id,
					"deploymentID": int64ToString(deployment.Id),
					"type":         "finished",
				}, *completedOn)
			}

			if completedOn != nil && startedOn != nil {
				deploymentStatusMetric.AddDuration(prometheus.Labels{
					"projectID":    project.Id,
					"deploymentID": int64ToString(deployment.Id),
					"type":         "jobDuration",
				}, completedOn.Sub(*startedOn))
			}
		}
	}

	callback <- func() {
		deploymentMetric.GaugeSet(m.prometheus.deployment)
		deploymentStatusMetric.GaugeSet(m.prometheus.deploymentStatus)
	}
}
