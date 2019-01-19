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
		repository        *prometheus.GaugeVec
		repositoryStats   *prometheus.GaugeVec
		repositoryCommits *prometheus.CounterVec
		repositoryPushes  *prometheus.CounterVec
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
	m.prometheus.repositoryPushes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "azure_devops_repository_pushes",
			Help: "Azure DevOps repository pushes",
		},
		[]string{
			"projectID",
			"repositoryID",
		},
	)

	prometheus.MustRegister(m.prometheus.repository)
	prometheus.MustRegister(m.prometheus.repositoryStats)
	prometheus.MustRegister(m.prometheus.repositoryCommits)
	prometheus.MustRegister(m.prometheus.repositoryPushes)
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
	fromTime := time.Now().Add(-*m.CollectorReference.GetScrapeTime())

	repositoryMetric := MetricCollectorList{}
	repositoryStatsMetric := MetricCollectorList{}
	repositoryCommitsMetric := MetricCollectorList{}
	repositoryPushesMetric := MetricCollectorList{}

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
	commitList, err := AzureDevopsClient.ListCommits(project.Name, repository.Id, fromTime)
	if err == nil {
		repositoryCommitsMetric.Add(prometheus.Labels{
			"projectID":    project.Id,
			"repositoryID": repository.Id,
		}, float64(commitList.Count))
	} else {
		Logger.Errorf("project[%v]call[ListCommits]: %v", project.Name, err)
	}

	// get pushes delta list
	pushList, err := AzureDevopsClient.ListPushes(project.Name, repository.Id, fromTime)
	if err == nil {
		repositoryPushesMetric.Add(prometheus.Labels{
			"projectID":    project.Id,
			"repositoryID": repository.Id,
		}, float64(pushList.Count))
	} else {
		Logger.Errorf("project[%v]call[ListCommits]: %v", project.Name, err)
	}

	callback <- func() {
		repositoryMetric.GaugeSet(m.prometheus.repository)
		repositoryStatsMetric.GaugeSet(m.prometheus.repositoryStats)
		repositoryCommitsMetric.CounterAdd(m.prometheus.repositoryCommits)
		repositoryPushesMetric.CounterAdd(m.prometheus.repositoryPushes)
	}
}
