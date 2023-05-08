package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"github.com/webdevops/go-common/utils/to"
	"go.uber.org/zap"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorPullRequest struct {
	collector.Processor

	prometheus struct {
		pullRequest       *prometheus.GaugeVec
		pullRequestStatus *prometheus.GaugeVec
		pullRequestLabel  *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorPullRequest) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

	m.prometheus.pullRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_pullrequest_info",
			Help: "Azure DevOps pullrequest",
		},
		[]string{
			"projectID",
			"repositoryID",
			"pullrequestID",
			"pullrequestTitle",
			"sourceBranch",
			"targetBranch",
			"status",
			"isDraft",
			"voteStatus",
			"creator",
		},
	)
	m.Collector.RegisterMetricList("pullRequest", m.prometheus.pullRequest, true)

	m.prometheus.pullRequestStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_pullrequest_status",
			Help: "Azure DevOps pullrequest status",
		},
		[]string{
			"projectID",
			"repositoryID",
			"pullrequestID",
			"type",
		},
	)
	m.Collector.RegisterMetricList("pullRequestStatus", m.prometheus.pullRequestStatus, true)

	m.prometheus.pullRequestLabel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_pullrequest_label",
			Help: "Azure DevOps pullrequest labels",
		},
		[]string{
			"projectID",
			"repositoryID",
			"pullrequestID",
			"label",
			"active",
		},
	)
	m.Collector.RegisterMetricList("pullRequestLabel", m.prometheus.pullRequestLabel, true)
}

func (m *MetricsCollectorPullRequest) Reset() {}

func (m *MetricsCollectorPullRequest) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))

		for _, repository := range project.RepositoryList.List {
			if repository.Disabled() {
				continue
			}

			repoLogger := projectLogger.With(zap.String("repository", repository.Name))
			m.collectPullRequests(ctx, repoLogger, callback, project, repository)
		}
	}
}

func (m *MetricsCollectorPullRequest) collectPullRequests(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	list, err := AzureDevopsClient.ListPullrequest(project.Id, repository.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	pullRequestMetric := m.Collector.GetMetricList("pullRequest")
	pullRequestStatusMetric := m.Collector.GetMetricList("pullRequestStatus")
	pullRequestLabelMetric := m.Collector.GetMetricList("pullRequestLabel")

	for _, pullRequest := range list.List {
		voteSummary := pullRequest.GetVoteSummary()

		pullRequestMetric.AddInfo(prometheus.Labels{
			"projectID":        project.Id,
			"repositoryID":     repository.Id,
			"pullrequestID":    int64ToString(pullRequest.Id),
			"pullrequestTitle": pullRequest.Title,
			"status":           pullRequest.Status,
			"voteStatus":       voteSummary.HumanizeString(),
			"creator":          pullRequest.CreatedBy.DisplayName,
			"isDraft":          to.BoolString(pullRequest.IsDraft),
			"sourceBranch":     pullRequest.SourceRefName,
			"targetBranch":     pullRequest.TargetRefName,
		})

		pullRequestStatusMetric.AddTime(prometheus.Labels{
			"projectID":     project.Id,
			"repositoryID":  repository.Id,
			"pullrequestID": int64ToString(pullRequest.Id),
			"type":          "created",
		}, pullRequest.CreationDate)

		for _, label := range pullRequest.Labels {
			pullRequestLabelMetric.AddInfo(prometheus.Labels{
				"projectID":     project.Id,
				"repositoryID":  repository.Id,
				"pullrequestID": int64ToString(pullRequest.Id),
				"label":         label.Name,
				"active":        to.BoolString(label.Active),
			})
		}
	}
}
