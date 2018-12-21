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
		repositoryStats *prometheus.GaugeVec
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


	m.prometheus.repositoryStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_repository_stats",
			Help: "Azure DevOps repository",
		},
		[]string{"projectID", "repositoryID", "type"},
	)

	prometheus.MustRegister(m.prometheus.repository)
	prometheus.MustRegister(m.prometheus.repositoryStats)
}

func (m *MetricsCollectorRepository) Reset() {
	m.prometheus.repository.Reset()
	m.prometheus.repositoryStats.Reset()
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
	repositoryStatsMetric := MetricCollectorList{}

	repositoryMetric.Add(
		prometheus.Labels{
			"projectID": project.Id,
			"repositoryID": repository.Id,
			"repositoryName": repository.Name,
		},
		1,
	)

	if repository.Size > 0 {
		repositoryStatsMetric.Add(
			prometheus.Labels{
				"projectID": project.Id,
				"repositoryID": repository.Id,
				"type": "size",
			},
			float64(repository.Size),
		)
	}

	// get commit delta list
	fromTime := time.Now().Add(- *m.CollectorReference.GetScrapeTime())
	commitList, err := AzureDevopsClient.ListCommits(project.Name, repository.Id, fromTime)
	if err == nil {
		repositoryStatsMetric.Add(
			prometheus.Labels{
				"projectID": project.Id,
				"repositoryID": repository.Id,
				"type": "commits",
			},
			float64(commitList.Count),
		)
	} else {
		ErrorLogger.Messsage("project[%v]call[ListCommits]: %v", project.Name, err)
	}

	callback <- func() {
		repositoryMetric.GaugeSet(m.prometheus.repository)
		repositoryStatsMetric.GaugeSet(m.prometheus.repositoryStats)
	}
}

