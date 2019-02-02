package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"github.com/prometheus/client_golang/prometheus"
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
			"voteStatus",
			"creator",
		},
	)

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

	prometheus.MustRegister(m.prometheus.pullRequest)
	prometheus.MustRegister(m.prometheus.pullRequestStatus)
	prometheus.MustRegister(m.prometheus.pullRequestLabel)
}

func (m *MetricsCollectorPullRequest) Reset() {
	m.prometheus.pullRequest.Reset()
	m.prometheus.pullRequestStatus.Reset()
	m.prometheus.pullRequestLabel.Reset()

}

func (m *MetricsCollectorPullRequest) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	for _, repository := range project.RepositoryList.List {
		m.collectPullRequests(ctx, callback, project, repository)
	}
}

func (m *MetricsCollectorPullRequest) collectPullRequests(ctx context.Context, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	list, err := AzureDevopsClient.ListPullrequest(project.Name, repository.Id)
	if err != nil {
		Logger.Errorf("project[%v]call[ListPullrequest] %v", project.Name, err)
		return
	}

	pullRequestMetric := MetricCollectorList{}
	pullRequestStatusMetric := MetricCollectorList{}
	pullRequestLabelMetric := MetricCollectorList{}

	for _, pullRequest := range list.List {
		voteSummary := pullRequest.GetVoteSummary()

		pullRequestMetric.AddInfo(prometheus.Labels{
			"projectID":        project.Id,
			"repositoryID":     repository.Id,
			"pullrequestID":    int64ToString(pullRequest.Id),
			"pullrequestTitle": pullRequest.Title,
			"status":           pullRequest.Status,
			"voteStatus":		voteSummary.HumanizeString(),
			"creator":          pullRequest.CreatedBy.DisplayName,
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
				"active":         boolToString(label.Active),
			})
		}
	}

	callback <- func() {
		pullRequestMetric.GaugeSet(m.prometheus.pullRequest)
		pullRequestStatusMetric.GaugeSet(m.prometheus.pullRequestStatus)
		pullRequestLabelMetric.GaugeSet(m.prometheus.pullRequestLabel)
	}
}
