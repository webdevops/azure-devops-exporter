package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	prometheusCommon "github.com/webdevops/go-common/prometheus"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorProject struct {
	CollectorProcessorProject

	prometheus struct {
		project    *prometheus.GaugeVec
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
		[]string{
			"projectID",
			"projectName",
		},
	)
	prometheus.MustRegister(m.prometheus.project)
}

func (m *MetricsCollectorProject) Reset() {
	m.prometheus.project.Reset()
}

func (m *MetricsCollectorProject) Collect(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	m.collectProject(ctx, logger, callback, project)
}

func (m *MetricsCollectorProject) collectProject(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	projectMetric := prometheusCommon.NewMetricsList()

	projectMetric.AddInfo(prometheus.Labels{
		"projectID":   project.Id,
		"projectName": project.Name,
	})

	callback <- func() {
		projectMetric.GaugeSet(m.prometheus.project)
	}
}
