package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

type MetricsCollectorRepository struct {
	CollectorProcessorProject

	prometheus struct {
		repository        *prometheus.GaugeVec
		repositoryStats   *prometheus.GaugeVec
		repositoryCommits *prometheus.CounterVec
	}
}

func (m *MetricsCollectorRepository) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.repository = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_repository_info",
			Help: "Azure DevOps repository",
		},
		[]string{
			"projectID",
			"repositoryID",
			"repositoryName",
		},
	)

	m.prometheus.repositoryStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_repository_stats",
			Help: "Azure DevOps repository",
		},
		[]string{
			"projectID",
			"repositoryID",
			"type",
		},
	)

	m.prometheus.repositoryCommits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "azure_devops_repository_commits",
			Help: "Azure DevOps repository commits",
		},
		[]string{
			"projectID",
			"repositoryID",
		},
	)

	prometheus.MustRegister(m.prometheus.repository)
	prometheus.MustRegister(m.prometheus.repositoryStats)
	prometheus.MustRegister(m.prometheus.repositoryCommits)
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
	repositoryCommitsMetric := MetricCollectorList{}

	repositoryMetric.AddInfo(prometheus.Labels{
		"projectID":      project.Id,
		"repositoryID":   repository.Id,
		"repositoryName": repository.Name,
	})

	if repository.Size > 0 {
		repositoryStatsMetric.Add(prometheus.Labels{
			"projectID":    project.Id,
			"repositoryID": repository.Id,
			"type":         "size",
		}, float64(repository.Size))
	}

	// get commit delta list
	fromTime := time.Now().Add(-*m.CollectorReference.GetScrapeTime())
	commitList, err := AzureDevopsClient.ListCommits(project.Name, repository.Id, fromTime)
	if err == nil {
		repositoryCommitsMetric.Add(prometheus.Labels{
			"projectID":    project.Id,
			"repositoryID": repository.Id,
		}, float64(commitList.Count))
	} else {
		LoggerError.Println(fmt.Sprintf("project[%v]call[ListCommits]: %v", project.Name, err))
	}

	callback <- func() {
		repositoryMetric.GaugeSet(m.prometheus.repository)
		repositoryStatsMetric.GaugeSet(m.prometheus.repositoryStats)
		repositoryCommitsMetric.CounterAdd(m.prometheus.repositoryCommits)
	}
}
