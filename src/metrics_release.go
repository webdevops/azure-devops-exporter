package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorRelease struct {
	CollectorProcessorProject

	prometheus struct {
		release           *prometheus.GaugeVec
		
		releaseDefinition *prometheus.GaugeVec
		releaseDefinitionEnvironment *prometheus.GaugeVec
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


	m.prometheus.releaseDefinitionEnvironment = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_definition_environment",
			Help: "Azure DevOps release definition environment",
		},
		[]string{"projectID", "releaseDefinitionID", "releaseDefinitionEnvironmentID", "releaseDefinitionEnvironmentName", "rank", "owner", "releaseID", "badgeUrl" },
	)

	prometheus.MustRegister(m.prometheus.release)
	prometheus.MustRegister(m.prometheus.releaseDefinition)
	prometheus.MustRegister(m.prometheus.releaseDefinitionEnvironment)
}

func (m *MetricsCollectorRelease) Reset() {
	m.prometheus.release.Reset()
	m.prometheus.releaseDefinition.Reset()
	m.prometheus.releaseDefinitionEnvironment.Reset()
}

func (m *MetricsCollectorRelease) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListReleaseDefinitions(project.Name)
	if err != nil {
		ErrorLogger.Messsage("project[%v]call[ListReleaseDefinitions]: %v", project.Name, err)
		return
	}

	releaseMetric := MetricCollectorList{}
	
	releaseDefinitionMetric := MetricCollectorList{}
	releaseDefinitionEnvironmentMetric := MetricCollectorList{}

	for _, releaseDefinition := range list.List {
		// --------------------------------------
		// Release definition
		infoLabels := prometheus.Labels{
			"projectID":              project.Id,
			"releaseDefinitionID":    int64ToString(releaseDefinition.Id),
			"releaseNameFormat":      releaseDefinition.ReleaseNameFormat,
			"releasedDefinitionName": releaseDefinition.Name,
			"url":                    releaseDefinition.Links.Web.Href,
		}
		releaseDefinitionMetric.Add(infoLabels, 1)
		
		for _, definitionEnvironments := range releaseDefinition.Environments {
			envLabels := prometheus.Labels{
				"projectID":                        project.Id,
				"releaseDefinitionID":              int64ToString(releaseDefinition.Id),
				"releaseDefinitionEnvironmentID":   int64ToString(definitionEnvironments.Id),
				"releaseDefinitionEnvironmentName": definitionEnvironments.Name,
				"rank":                             int64ToString(definitionEnvironments.Rank),
				"owner":                            definitionEnvironments.Owner.DisplayName,
				"releaseID":                        int64ToString(definitionEnvironments.CurrentRelease.Id),
				"badgeUrl":                         definitionEnvironments.BadgeUrl,
			}
			releaseDefinitionEnvironmentMetric.Add(envLabels, 1)
		}

		// --------------------------------------
		// Releases
		
		releaseList, err := AzureDevopsClient.ListReleases(project.Name, releaseDefinition.Id)
		if err != nil {
			ErrorLogger.Messsage("project[%v]call[ListReleases]: %v", project.Name, err)
			return
		}

		for _, release := range releaseList.List {
			infoLabels := prometheus.Labels{
				"projectID":           project.Id,
				"releaseID":           int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"requestedBy":         release.RequestedBy.DisplayName,
				"releasedName":        release.Name,
				"status":              release.Status,
				"reason":              release.Reason,
				"result":              boolToString(release.Result),
				"url":                 release.Links.Web.Href,
			}

			releaseMetric.Add(infoLabels, 1)
		}
	}

	callback <- func() {
		releaseDefinitionMetric.GaugeSet(m.prometheus.releaseDefinition)
		releaseDefinitionEnvironmentMetric.GaugeSet(m.prometheus.releaseDefinitionEnvironment)

		releaseMetric.GaugeSet(m.prometheus.release)
	}
}
