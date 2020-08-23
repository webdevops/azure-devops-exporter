package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	prometheusCommon "github.com/webdevops/go-prometheus-common"
	"sync"
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
	prometheus.MustRegister(m.prometheus.repository)

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
	prometheus.MustRegister(m.prometheus.repositoryStats)

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
	prometheus.MustRegister(m.prometheus.repositoryCommits)

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
	prometheus.MustRegister(m.prometheus.repositoryPushes)
}

func (m *MetricsCollectorRepository) Reset() {
	m.prometheus.repository.Reset()
	m.prometheus.repositoryStats.Reset()
}

func (m *MetricsCollectorRepository) Collect(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	wg := sync.WaitGroup{}

	for _, repository := range project.RepositoryList.List {
		wg.Add(1)
		go func(ctx context.Context, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
			defer wg.Done()
			contextLogger := logger.WithField("repository", repository.Name)
			m.collectRepository(ctx, contextLogger, callback, project, repository)
		}(ctx, callback, project, repository)
	}

	wg.Wait()
}

func (m *MetricsCollectorRepository) collectRepository(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	fromTime := *m.CollectorReference.collectionLastTime

	repositoryMetric := prometheusCommon.NewMetricsList()
	repositoryStatsMetric := prometheusCommon.NewMetricsList()
	repositoryCommitsMetric := prometheusCommon.NewMetricsList()
	repositoryPushesMetric := prometheusCommon.NewMetricsList()

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
	commitList, err := AzureDevopsClient.ListCommits(project.Id, repository.Id, fromTime)
	if err == nil {
		repositoryCommitsMetric.Add(prometheus.Labels{
			"projectID":    project.Id,
			"repositoryID": repository.Id,
		}, float64(commitList.Count))
	} else {
		logger.Error(err)
	}

	// get pushes delta list
	pushList, err := AzureDevopsClient.ListPushes(project.Id, repository.Id, fromTime)
	if err == nil {
		repositoryPushesMetric.Add(prometheus.Labels{
			"projectID":    project.Id,
			"repositoryID": repository.Id,
		}, float64(pushList.Count))
	} else {
		logger.Error(err)
	}

	callback <- func() {
		repositoryMetric.GaugeSet(m.prometheus.repository)
		repositoryStatsMetric.GaugeSet(m.prometheus.repositoryStats)
		repositoryCommitsMetric.CounterAdd(m.prometheus.repositoryCommits)
		repositoryPushesMetric.CounterAdd(m.prometheus.repositoryPushes)
	}
}
