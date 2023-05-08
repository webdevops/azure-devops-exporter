package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/remeh/sizedwaitgroup"
	"github.com/webdevops/go-common/prometheus/collector"
	"go.uber.org/zap"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorRepository struct {
	collector.Processor

	prometheus struct {
		repository        *prometheus.GaugeVec
		repositoryStats   *prometheus.GaugeVec
		repositoryCommits *prometheus.CounterVec
		repositoryPushes  *prometheus.CounterVec
	}
}

func (m *MetricsCollectorRepository) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

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
	m.Collector.RegisterMetricList("repository", m.prometheus.repository, true)

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
	m.Collector.RegisterMetricList("repositoryStats", m.prometheus.repositoryStats, true)

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
	m.Collector.RegisterMetricList("repositoryCommits", m.prometheus.repositoryCommits, false)

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
	m.Collector.RegisterMetricList("repositoryPushes", m.prometheus.repositoryPushes, false)
}

func (m *MetricsCollectorRepository) Reset() {}

func (m *MetricsCollectorRepository) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))

		wg := sizedwaitgroup.New(5)
		for _, repository := range project.RepositoryList.List {
			if repository.Disabled() {
				continue
			}

			wg.Add()
			go func(ctx context.Context, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
				defer wg.Done()
				repositoryLogger := projectLogger.With(zap.String("repository", repository.Name))
				m.collectRepository(ctx, repositoryLogger, callback, project, repository)
			}(ctx, callback, project, repository)
		}
		wg.Wait()
	}
}

func (m *MetricsCollectorRepository) collectRepository(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	fromTime := time.Now().Add(-*m.Collector.GetScapeTime())
	if val := m.Collector.GetLastScapeTime(); val != nil {
		fromTime = *val
	}

	repositoryMetric := m.Collector.GetMetricList("repository")
	repositoryStatsMetric := m.Collector.GetMetricList("repositoryStats")
	repositoryCommitsMetric := m.Collector.GetMetricList("repositoryCommits")
	repositoryPushesMetric := m.Collector.GetMetricList("repositoryPushes")

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
}
