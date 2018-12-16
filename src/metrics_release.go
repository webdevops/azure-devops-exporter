package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
)

type MetricsCollectorRelease struct {
	CollectorProcessorProject

	prometheus struct {
		release *prometheus.GaugeVec
		releaseDefinition *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorRelease) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.release = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_info",
			Help: "Azure DevOps release",
		},
		[]string{"projectID", "releaseID", "releaseDefinitionID", "requestedBy", "releasedName", "status", "reason", "result", "url"},
	)

	m.prometheus.releaseDefinition = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_definition",
			Help: "Azure DevOps release definition",
		},
		[]string{"projectID", "releaseDefinitionID", "releaseNameFormat", "releasedDefinitionName", "url"},
	)

	prometheus.MustRegister(m.prometheus.release)
	prometheus.MustRegister(m.prometheus.releaseDefinition)
}

func (m *MetricsCollectorRelease) Reset() {
	m.prometheus.release.Reset()
	m.prometheus.releaseDefinition.Reset()
}

func (m *MetricsCollectorRelease) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListReleaseDefinitions(project.Name)
	if err != nil {
		ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
		return
	}

	releaseMetric := MetricCollectorList{}
	releaseDefinitionMetric := MetricCollectorList{}

	for _, releaseDefinition := range list.List {
		infoLabels := prometheus.Labels{
			"projectID": project.Id,
			"releaseDefinitionID": int64ToString(releaseDefinition.Id),
			"releaseNameFormat": releaseDefinition.ReleaseNameFormat,
			"releasedDefinitionName": releaseDefinition.Name,
			"url": releaseDefinition.Links.Web.Href,
		}
		releaseDefinitionMetric.Add(infoLabels, 1)


		releaseList, err := AzureDevopsClient.ListReleases(project.Name, releaseDefinition.Id)
		if err != nil {
			ErrorLogger.Messsage("project[%v]: %v", project.Name, err)
			return
		}

		for _, release := range releaseList.List {
			infoLabels := prometheus.Labels{
				"projectID": project.Id,
				"releaseID": int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"requestedBy": release.RequestedBy.DisplayName,
				"releasedName": release.Name,
				"status": release.Status,
				"reason": release.Reason,
				"result": release.Result,
				"url": release.Links.Web.Href,
			}

			releaseMetric.Add(infoLabels, 1)
		}
	}

	callback <- func() {
		releaseMetric.GaugeSet(m.prometheus.release)
		releaseDefinitionMetric.GaugeSet(m.prometheus.releaseDefinition)
	}
}
