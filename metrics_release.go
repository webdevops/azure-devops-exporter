package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"github.com/webdevops/go-common/utils/to"
	"go.uber.org/zap"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorRelease struct {
	collector.Processor

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

func (m *MetricsCollectorRelease) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

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
	m.Collector.RegisterMetricList("release", m.prometheus.release, true)

	m.prometheus.releaseArtifact = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_artifact",
			Help: "Azure DevOps release",
		},
		[]string{
			"projectID",
			"releaseID",
			"releaseDefinitionID",
			"sourceId",
			"repositoryID",
			"branch",
			"type",
			"alias",
			"version",
		},
	)
	m.Collector.RegisterMetricList("releaseArtifact", m.prometheus.releaseArtifact, true)

	m.prometheus.releaseEnvironment = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_environment",
			Help: "Azure DevOps release environment",
		},
		[]string{
			"projectID",
			"releaseID",
			"releaseDefinitionID",
			"environmentID",
			"environmentName",
			"status",
			"triggerReason",
			"rank",
		},
	)
	m.Collector.RegisterMetricList("releaseEnvironment", m.prometheus.releaseEnvironment, true)

	m.prometheus.releaseEnvironmentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_environment_status",
			Help: "Azure DevOps release environment status",
		},
		[]string{
			"projectID",
			"releaseID",
			"releaseDefinitionID",
			"environmentID",
			"type",
		},
	)
	m.Collector.RegisterMetricList("releaseEnvironmentStatus", m.prometheus.releaseEnvironmentStatus, true)

	m.prometheus.releaseEnvironmentApproval = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_release_approval",
			Help: "Azure DevOps release approval",
		},
		[]string{
			"projectID",
			"releaseID",
			"releaseDefinitionID",
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
	m.Collector.RegisterMetricList("releaseEnvironmentApproval", m.prometheus.releaseEnvironmentApproval, true)

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
	m.Collector.RegisterMetricList("releaseDefinition", m.prometheus.releaseDefinition, true)

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
	m.Collector.RegisterMetricList("releaseDefinitionEnvironment", m.prometheus.releaseDefinitionEnvironment, true)
}

func (m *MetricsCollectorRelease) Reset() {}

func (m *MetricsCollectorRelease) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))
		m.collectReleases(ctx, projectLogger, callback, project)
	}
}

func (m *MetricsCollectorRelease) collectReleases(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListReleaseDefinitions(project.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	releaseDefinitionMetric := m.Collector.GetMetricList("releaseDefinition")
	releaseDefinitionEnvironmentMetric := m.Collector.GetMetricList("releaseDefinitionEnvironment")

	releaseMetric := m.Collector.GetMetricList("release")
	releaseArtifactMetric := m.Collector.GetMetricList("releaseArtifact")
	releaseEnvironmentMetric := m.Collector.GetMetricList("releaseEnvironment")
	releaseEnvironmentApprovalMetric := m.Collector.GetMetricList("releaseEnvironmentApproval")
	releaseEnvironmentStatusMetric := m.Collector.GetMetricList("releaseEnvironmentStatus")

	for _, releaseDefinition := range list.List {
		// --------------------------------------
		// Release definition
		releaseDefinitionMetric.AddInfo(prometheus.Labels{
			"projectID":             project.Id,
			"releaseDefinitionID":   int64ToString(releaseDefinition.Id),
			"releaseNameFormat":     releaseDefinition.ReleaseNameFormat,
			"releaseDefinitionName": releaseDefinition.Name,
			"path":                  releaseDefinition.Path,
			"url":                   releaseDefinition.Links.Web.Href,
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
	}

	// --------------------------------------
	// Releases
	minTime := time.Now().Add(-Opts.Limit.ReleaseHistoryDuration)

	releaseList, err := AzureDevopsClient.ListReleaseHistory(project.Id, minTime)
	if err != nil {
		logger.Error(err)
		return
	}

	for _, release := range releaseList.List {
		releaseMetric.AddInfo(prometheus.Labels{
			"projectID":           project.Id,
			"releaseID":           int64ToString(release.Id),
			"releaseDefinitionID": int64ToString(release.Definition.Id),
			"requestedBy":         release.RequestedBy.DisplayName,
			"releaseName":         release.Name,
			"status":              release.Status,
			"reason":              release.Reason,
			"result":              to.BoolString(release.Result),
			"url":                 release.Links.Web.Href,
		})

		for _, artifact := range release.Artifacts {
			releaseArtifactMetric.AddInfo(prometheus.Labels{
				"projectID":           project.Id,
				"releaseID":           int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"sourceId":            artifact.SourceId,
				"repositoryID":        artifact.DefinitionReference.Repository.Name,
				"branch":              artifact.DefinitionReference.Branch.Name,
				"type":                artifact.Type,
				"alias":               artifact.Alias,
				"version":             artifact.DefinitionReference.Version.Name,
			})
		}

		for _, environment := range release.Environments {
			releaseEnvironmentMetric.AddInfo(prometheus.Labels{
				"projectID":           project.Id,
				"releaseID":           int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"environmentID":       int64ToString(environment.DefinitionEnvironmentId),
				"environmentName":     environment.Name,
				"status":              environment.Status,
				"triggerReason":       environment.TriggerReason,
				"rank":                int64ToString(environment.Rank),
			})

			releaseEnvironmentStatusMetric.AddBool(prometheus.Labels{
				"projectID":           project.Id,
				"releaseID":           int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"environmentID":       int64ToString(environment.DefinitionEnvironmentId),
				"type":                "succeeded",
			}, environment.Status == "succeeded")

			releaseEnvironmentStatusMetric.AddTime(prometheus.Labels{
				"projectID":           project.Id,
				"releaseID":           int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"environmentID":       int64ToString(environment.DefinitionEnvironmentId),
				"type":                "created",
			}, environment.CreatedOn)

			releaseEnvironmentStatusMetric.AddIfNotZero(prometheus.Labels{
				"projectID":           project.Id,
				"releaseID":           int64ToString(release.Id),
				"releaseDefinitionID": int64ToString(release.Definition.Id),
				"environmentID":       int64ToString(environment.DefinitionEnvironmentId),
				"type":                "jobDuration",
			}, environment.TimeToDeploy*60)

			for _, approval := range environment.PreDeployApprovals {
				// skip automated approvals
				if approval.IsAutomated {
					continue
				}

				releaseEnvironmentApprovalMetric.AddTime(prometheus.Labels{
					"projectID":           project.Id,
					"releaseID":           int64ToString(release.Id),
					"releaseDefinitionID": int64ToString(release.Definition.Id),
					"environmentID":       int64ToString(environment.DefinitionEnvironmentId),
					"approvalType":        approval.ApprovalType,
					"status":              approval.Status,
					"isAutomated":         to.BoolString(approval.IsAutomated),
					"trialNumber":         int64ToString(approval.TrialNumber),
					"attempt":             int64ToString(approval.Attempt),
					"rank":                int64ToString(approval.Rank),
					"approver":            approval.Approver.DisplayName,
					"approvedBy":          approval.ApprovedBy.DisplayName,
				}, approval.CreatedOn)
			}

			for _, approval := range environment.PostDeployApprovals {
				// skip automated approvals
				if approval.IsAutomated {
					continue
				}

				releaseEnvironmentApprovalMetric.AddTime(prometheus.Labels{
					"projectID":           project.Id,
					"releaseID":           int64ToString(release.Id),
					"releaseDefinitionID": int64ToString(release.Definition.Id),
					"environmentID":       int64ToString(environment.DefinitionEnvironmentId),
					"approvalType":        approval.ApprovalType,
					"status":              approval.Status,
					"isAutomated":         to.BoolString(approval.IsAutomated),
					"trialNumber":         int64ToString(approval.TrialNumber),
					"attempt":             int64ToString(approval.Attempt),
					"rank":                int64ToString(approval.Rank),
					"approver":            approval.Approver.DisplayName,
					"approvedBy":          approval.ApprovedBy.DisplayName,
				}, approval.CreatedOn)
			}
		}
	}
}
