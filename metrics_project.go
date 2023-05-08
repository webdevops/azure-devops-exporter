package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"go.uber.org/zap"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorProject struct {
	collector.Processor

	prometheus struct {
		project    *prometheus.GaugeVec
		repository *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorProject) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

	m.prometheus.project = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_project_info",
			Help: "Azure DevOps project",
		},
		[]string{
			"projectID",
			"projectName",
		},
	)
	m.Collector.RegisterMetricList("project", m.prometheus.project, true)
}

func (m *MetricsCollectorProject) Reset() {}

func (m *MetricsCollectorProject) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))
		m.collectProject(ctx, projectLogger, callback, project)
	}
}

func (m *MetricsCollectorProject) collectProject(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), project devopsClient.Project) {
	projectMetric := m.Collector.GetMetricList("project")

	projectMetric.AddInfo(prometheus.Labels{
		"projectID":   project.Id,
		"projectName": project.Name,
	})
}
