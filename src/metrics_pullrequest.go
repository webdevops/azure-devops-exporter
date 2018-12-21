package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
)

type MetricsCollectorPullRequest struct {
	CollectorProcessorProject

	prometheus struct {
		pullRequest *prometheus.GaugeVec
		pullRequestStatus *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorPullRequest) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.pullRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_pullrequest_info",
			Help: "Azure DevOps pullrequest",
		},
		[]string{"projectID", "repositoryID", "pullrequestID", "pullrequestTitle", "sourceBranch", "targetBranch", "status", "creator"},
	)

	m.prometheus.pullRequestStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_pullrequest_status",
			Help: "Azure DevOps pullrequest",
		},
		[]string{"projectID", "repositoryID", "pullrequestID", "type"},
	)

	prometheus.MustRegister(m.prometheus.pullRequest)
	prometheus.MustRegister(m.prometheus.pullRequestStatus)
}

func (m *MetricsCollectorPullRequest) Reset() {
	m.prometheus.pullRequest.Reset()
	m.prometheus.pullRequestStatus.Reset()

}

func (m *MetricsCollectorPullRequest) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	for _, repository := range project.RepositoryList.List {
		m.collectPullRequests(ctx, callback, project, repository)
	}
}

func (m *MetricsCollectorPullRequest) collectPullRequests(ctx context.Context, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	list, err := AzureDevopsClient.ListPullrequest(project.Name, repository.Id)
	if err != nil {
		ErrorLogger.Messsage("project[%v]call[ListPullrequest] %v", project.Name, err)
		return
	}

	pullRequestMetric := MetricCollectorList{}
	pullRequestStatusMetric := MetricCollectorList{}

	for _, pullRequest := range list.List {
		infoLabels := prometheus.Labels{
			"projectID":     project.Id,
			"repositoryID":  repository.Id,
			"pullrequestID": int64ToString(pullRequest.Id),
			"pullrequestTitle": pullRequest.Title,
			"status": pullRequest.Status,
			"creator": pullRequest.CreatedBy.DisplayName,
			"sourceBranch": pullRequest.SourceRefName,
			"targetBranch": pullRequest.TargetRefName,
		}
		pullRequestMetric.Add(infoLabels, 1)

		statusCreatedLabels := prometheus.Labels{
			"projectID":     project.Id,
			"repositoryID":  repository.Id,
			"pullrequestID": int64ToString(pullRequest.Id),
			"type": "created",
		}
		statusCreateValue := float64(pullRequest.CreationDate.Unix())
		if statusCreateValue > 0 {
			pullRequestStatusMetric.Add(statusCreatedLabels, statusCreateValue)
		}
	}

	callback <- func() {
		pullRequestMetric.GaugeSet(m.prometheus.pullRequest)
		pullRequestStatusMetric.GaugeSet(m.prometheus.pullRequestStatus)
	}
}
