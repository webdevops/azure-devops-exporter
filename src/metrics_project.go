package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
)

type MetricsCollectorProject struct {
	CollectorProcessorProject

	prometheus struct {
		project *prometheus.GaugeVec
		repository *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorProject) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.project = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_project_info",
			Help: "Azure DevOps project",
		},
		[]string{"projectID", "projectName"},
	)

	m.prometheus.repository = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_repository_info",
			Help: "Azure DevOps repository",
		},
		[]string{"projectID", "repositoryID", "repositoryName"},
	)

	prometheus.MustRegister(m.prometheus.project)
	prometheus.MustRegister(m.prometheus.repository)
}

func (m *MetricsCollectorProject) Reset() {
	m.prometheus.project.Reset()
	m.prometheus.repository.Reset()
}

func (m *MetricsCollectorProject) Collect(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	m.collectProject(ctx, callback, project)

	for _, repository := range project.RepositoryList.List {
		m.collectRepository(ctx, callback, project, repository)
	}
}

func (m *MetricsCollectorProject) collectProject(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	projectMetric := MetricCollectorList{}

	infoLabels := prometheus.Labels{
		"projectID": project.Id,
		"projectName": project.Name,
	}
	projectMetric.Add(infoLabels, 1)

	callback <- func() {
		projectMetric.GaugeSet(m.prometheus.project)
	}
}


func (m *MetricsCollectorProject) collectRepository(ctx context.Context, callback chan<- func(), project devopsClient.Project, repository devopsClient.Repository) {
	repositoryMetric := MetricCollectorList{}

	infoLabels := prometheus.Labels{
		"projectID": project.Id,
		"repositoryID": repository.Id,
		"repositoryName": repository.Name,
	}
	repositoryMetric.Add(infoLabels, 1)

	callback <- func() {
		repositoryMetric.GaugeSet(m.prometheus.repository)
	}
}

