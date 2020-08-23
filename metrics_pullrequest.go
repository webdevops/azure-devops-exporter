package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	prometheusCommon "github.com/webdevops/go-prometheus-common"
)

type MetricsCollectorPullRequest struct {
	CollectorProcessorProject

	prometheus struct {
		pullRequest       *prometheus.GaugeVec
		pullRequestStatus *prometheus.GaugeVec
		pullRequestLabel  *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorPullRequest) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

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
	prometheus.MustRegister(m.prometheus.pullRequest)

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
	prometheus.MustRegister(m.prometheus.pullRequestStatus)

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
	prometheus.MustRegister(m.prometheus.pullRequestLabel)
}

func (m *MetricsCollectorPullRequest) Reset() {
	m.prometheus.pullRequest.Reset()
	m.prometheus.pullRequestStatus.Reset()
	m.prometheus.pullRequestLabel.Reset()

}

func (m *MetricsCollectorPullRequest) Collect(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	for _, repository := range project.RepositoryList.List {
		contextLogger := logger.WithField("repository", repository.Name)
		m.collectPullRequests(ctx, contextLogger, callback, project, repository)
	}
}

func (m *MetricsCollectorPullRequest) collectPullRequests(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	list, err := AzureDevopsClient.ListPullrequest(project.Id, repository.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	pullRequestMetric := prometheusCommon.NewMetricsList()
	pullRequestStatusMetric := prometheusCommon.NewMetricsList()
	pullRequestLabelMetric := prometheusCommon.NewMetricsList()

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
			"isDraft":          boolToString(pullRequest.IsDraft),
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
				"active":        boolToString(label.Active),
			})
		}
	}

	callback <- func() {
		pullRequestMetric.GaugeSet(m.prometheus.pullRequest)
		pullRequestStatusMetric.GaugeSet(m.prometheus.pullRequestStatus)
		pullRequestLabelMetric.GaugeSet(m.prometheus.pullRequestLabel)
	}
}
