package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorRelease struct {
	CollectorProcessorProject

	prometheus struct {
		release                    *prometheus.GaugeVec
		releaseArtifact            *prometheus.GaugeVec
		releaseEnvironment         *prometheus.GaugeVec
		releaseEnvironmentApproval *prometheus.GaugeVec
		releaseEnvironmentStatus   *prometheus.GaugeVec

		releaseDefinition            *prometheus.GaugeVec
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
		[]string{
			"projectID",
			"releaseID",
			"releaseDefinitionID",
			"requestedBy",
			"releaseName",
			"status",
			"reason",
			"result",
			"url",
		},
	)

	m.prometheus.releaseArtifact = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_artifact",
			Help: "Azure DevOps release",
		},
		[]string{
			"projectID",
			"releaseID",
			"sourceId",
			"repositoryID",
			"branch",
			"type",
			"alias",
			"version",
		},
	)

	m.prometheus.releaseEnvironment = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_environment",
			Help: "Azure DevOps release environment",
		},
		[]string{
			"projectID",
			"releaseID",
			"environmentID",
			"environmentName",
			"status",
			"triggerReason",
			"rank",
		},
	)

	m.prometheus.releaseEnvironmentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_environment_status",
			Help: "Azure DevOps release environment status",
		},
		[]string{
			"projectID",
			"releaseID",
			"environmentID",
			"type",
		},
	)

	m.prometheus.releaseEnvironmentApproval = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_approval",
			Help: "Azure DevOps release approval",
		},
		[]string{
			"projectID",
			"releaseID",
			"environmentID",
			"approvalType",
			"status",
			"isAutomated",
			"trialNumber",
			"attempt",
			"rank",
			"approver",
			"approvedBy",
		},
	)

	m.prometheus.releaseDefinition = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_definition_info",
			Help: "Azure DevOps release definition",
		},
		[]string{
			"projectID",
			"releaseDefinitionID",
			"releaseNameFormat",
			"releaseDefinitionName",
			"path",
			"url",
		},
	)

	m.prometheus.releaseDefinitionEnvironment = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_definition_environment",
			Help: "Azure DevOps release definition environment",
		},
		[]string{
			"projectID",
			"releaseDefinitionID",
			"environmentID",
			"environmentName",
			"rank",
			"owner",
			"releaseID",
			"badgeUrl",
		},
	)

	prometheus.MustRegister(m.prometheus.release)
	prometheus.MustRegister(m.prometheus.releaseArtifact)
	prometheus.MustRegister(m.prometheus.releaseEnvironment)
	prometheus.MustRegister(m.prometheus.releaseEnvironmentApproval)
	prometheus.MustRegister(m.prometheus.releaseEnvironmentStatus)
	prometheus.MustRegister(m.prometheus.releaseDefinition)
	prometheus.MustRegister(m.prometheus.releaseDefinitionEnvironment)
}

func (m *MetricsCollectorRelease) Reset() {
	m.prometheus.release.Reset()
	m.prometheus.releaseArtifact.Reset()
	m.prometheus.releaseEnvironment.Reset()
	m.prometheus.releaseEnvironmentApproval.Reset()
	m.prometheus.releaseEnvironmentStatus.Reset()

	m.prometheus.releaseDefinition.Reset()
	m.prometheus.releaseDefinitionEnvironment.Reset()
}

func (m *MetricsCollectorRelease) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListReleaseDefinitions(project.Name)
	if err != nil {
		Logger.Errorf("project[%v]call[ListReleaseDefinitions]: %v", project.Name, err)
		return
	}

	releaseDefinitionMetric := MetricCollectorList{}
	releaseDefinitionEnvironmentMetric := MetricCollectorList{}

	releaseMetric := MetricCollectorList{}
	releaseArtifactMetric := MetricCollectorList{}
	releaseEnvironmentMetric := MetricCollectorList{}
	releaseEnvironmentApprovalMetric := MetricCollectorList{}
	releaseEnvironmentStatusMetric := MetricCollectorList{}

	for _, releaseDefinition := range list.List {
		// --------------------------------------
		// Release definition
		releaseDefinitionMetric.AddInfo(prometheus.Labels{
			"projectID":              project.Id,
			"releaseDefinitionID":    int64ToString(releaseDefinition.Id),
			"releaseNameFormat":      releaseDefinition.ReleaseNameFormat,
			"releaseDefinitionName": releaseDefinition.Name,
			"path":                   releaseDefinition.Path,
			"url":                    releaseDefinition.Links.Web.Href,
		})

		for _, environment := range releaseDefinition.Environments {
			releaseDefinitionEnvironmentMetric.AddInfo(prometheus.Labels{
				"projectID":           project.Id,
				"releaseDefinitionID": int64ToString(releaseDefinition.Id),
				"environmentID":       int64ToString(environment.Id),
				"environmentName":     environment.Name,
				"rank":                int64ToString(environment.Rank),
				"owner":               environment.Owner.DisplayName,
				"releaseID":           int64ToString(environment.CurrentRelease.Id),
				"badgeUrl":            environment.BadgeUrl,
			})
		}

		// --------------------------------------
		// Releases

		releaseList, err := AzureDevopsClient.ListReleases(project.Name, releaseDefinition.Id)
		if err != nil {
			Logger.Errorf("project[%v]call[ListReleases]: %v", project.Name, err)
			return
		}

		for _, release := range releaseList.List {
			releaseMetric.AddInfo(prometheus.Labels{
				"projectID":           project.Id,
				"releaseID":           int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"requestedBy":         release.RequestedBy.DisplayName,
				"releaseName":        release.Name,
				"status":              release.Status,
				"reason":              release.Reason,
				"result":              boolToString(release.Result),
				"url":                 release.Links.Web.Href,
			})

			for _, artifact := range release.Artifacts {
				releaseArtifactMetric.AddInfo(prometheus.Labels{
					"projectID":       project.Id,
					"releaseID":       int64ToString(release.Id),
					"sourceId":        artifact.SourceId,
					"repositoryID":    artifact.DefinitionReference.Repository.Name,
					"branch":          artifact.DefinitionReference.Branch.Name,
					"type":            artifact.Type,
					"alias":           artifact.Alias,
					"version":         artifact.DefinitionReference.Version.Name,
				})
			}

			for _, environment := range release.Environments {
				releaseEnvironmentMetric.AddInfo(prometheus.Labels{
					"projectID":       project.Id,
					"releaseID":       int64ToString(release.Id),
					"environmentID":   int64ToString(environment.DefinitionEnvironmentId),
					"environmentName": environment.Name,
					"status":          environment.Status,
					"triggerReason":   environment.TriggerReason,
					"rank":            int64ToString(environment.Rank),
				})

				releaseEnvironmentStatusMetric.AddTime(prometheus.Labels{
					"projectID":     project.Id,
					"releaseID":     int64ToString(release.Id),
					"environmentID": int64ToString(environment.DefinitionEnvironmentId),
					"type":          "created",
				}, environment.CreatedOn)

				releaseEnvironmentStatusMetric.Add(prometheus.Labels{
					"projectID":     project.Id,
					"releaseID":     int64ToString(release.Id),
					"environmentID": int64ToString(environment.DefinitionEnvironmentId),
					"type":          "jobDuration",
				}, environment.TimeToDeploy)

				for _, approval := range environment.PreDeployApprovals {
					// skip automated approvals
					if approval.IsAutomated {
						continue
					}

					releaseEnvironmentApprovalMetric.AddTime(prometheus.Labels{
						"projectID":     project.Id,
						"releaseID":     int64ToString(release.Id),
						"environmentID": int64ToString(environment.DefinitionEnvironmentId),
						"approvalType":  approval.ApprovalType,
						"status":        approval.Status,
						"isAutomated":   boolToString(approval.IsAutomated),
						"trialNumber":   int64ToString(approval.TrialNumber),
						"attempt":       int64ToString(approval.Attempt),
						"rank":          int64ToString(approval.Rank),
						"approver":      approval.Approver.DisplayName,
						"approvedBy":    approval.ApprovedBy.DisplayName,
					}, approval.CreatedOn)
				}

				for _, approval := range environment.PostDeployApprovals {
					// skip automated approvals
					if approval.IsAutomated {
						continue
					}

					releaseEnvironmentApprovalMetric.AddTime(prometheus.Labels{
						"projectID":     project.Id,
						"releaseID":     int64ToString(release.Id),
						"environmentID": int64ToString(environment.DefinitionEnvironmentId),
						"approvalType":  approval.ApprovalType,
						"status":        approval.Status,
						"isAutomated":   boolToString(approval.IsAutomated),
						"trialNumber":   int64ToString(approval.TrialNumber),
						"attempt":       int64ToString(approval.Attempt),
						"rank":          int64ToString(approval.Rank),
						"approver":      approval.Approver.DisplayName,
						"approvedBy":    approval.ApprovedBy.DisplayName,
					}, approval.CreatedOn)
				}
			}
		}
	}

	callback <- func() {
		releaseDefinitionMetric.GaugeSet(m.prometheus.releaseDefinition)
		releaseDefinitionEnvironmentMetric.GaugeSet(m.prometheus.releaseDefinitionEnvironment)

		releaseMetric.GaugeSet(m.prometheus.release)
		releaseArtifactMetric.GaugeSet(m.prometheus.releaseArtifact)
		releaseEnvironmentMetric.GaugeSet(m.prometheus.releaseEnvironment)
		releaseEnvironmentApprovalMetric.GaugeSet(m.prometheus.releaseEnvironmentApproval)
		releaseEnvironmentStatusMetric.GaugeSet(m.prometheus.releaseEnvironmentStatus)
	}
}
