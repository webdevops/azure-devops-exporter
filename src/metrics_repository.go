package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

type MetricsCollectorRepository struct {
	CollectorProcessorProject

	prometheus struct {
		repository *prometheus.GaugeVec
		repositoryCommits *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorRepository) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.repository = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_repository_info",
			Help: "Azure DevOps repository",
		},
		[]string{"projectID", "repositoryID", "repositoryName"},
	)


	m.prometheus.repositoryCommits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_repository_commits",
			Help: "Azure DevOps repository",
		},
		[]string{"projectID", "repositoryID"},
	)

	prometheus.MustRegister(m.prometheus.repository)
	prometheus.MustRegister(m.prometheus.repositoryCommits)
}

func (m *MetricsCollectorRepository) Reset() {
	m.prometheus.repository.Reset()
	m.prometheus.repositoryCommits.Reset()
}

func (m *MetricsCollectorRepository) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	wg := sync.WaitGroup{}

	for _, repository := range project.RepositoryList.List {
		wg.Add(1)
		go func(ctx context.Context, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
			defer wg.Done()
			m.collectRepository(ctx, callback, project, repository)
		}(ctx, callback, project, repository)
	}

	wg.Wait()
}

func (m *MetricsCollectorRepository) collectRepository(ctx context.Context, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	repositoryMetric := MetricCollectorList{}
	repositoryCommitsMetric := MetricCollectorList{}

	repositoryMetric.Add(
		prometheus.Labels{
			"projectID": project.Id,
			"repositoryID": repository.Id,
			"repositoryName": repository.Name,
		},
		1,
	)

	callback <- func() {
		repositoryMetric.GaugeSet(m.prometheus.repository)
	}

	fromTime := time.Now().Add(- *m.CollectorReference.GetScrapeTime())
	commitList, err := AzureDevopsClient.ListCommits(project.Name, repository.Id, fromTime)

	if err != nil {
		ErrorLogger.Messsage("project[%v]call[ListCommits]: %v", project.Name, err)
		return
	}

	repositoryCommitsMetric.Add(
		prometheus.Labels{
			"projectID": project.Id,
			"repositoryID": repository.Id,
		},
		float64(commitList.Count),
	)

	callback <- func() {
		repositoryCommitsMetric.GaugeSet(m.prometheus.repositoryCommits)
	}
}

