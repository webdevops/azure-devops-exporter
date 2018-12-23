package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorGeneral struct {
	CollectorProcessorGeneral

	prometheus struct {
		stats *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorGeneral) Setup(collector *CollectorGeneral) {
	m.CollectorReference = collector

	m.prometheus.stats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_stats",
			Help: "Azure DevOps statistics",
		},
		[]string{"type"},
	)

	prometheus.MustRegister(m.prometheus.stats)
}

func (m *MetricsCollectorGeneral) Reset() {
	m.prometheus.stats.Reset()
}

func (m *MetricsCollectorGeneral) Collect(ctx context.Context, callback chan<- func()) {
	statsMetrics := MetricCollectorList{}

	statsMetrics.Add(prometheus.Labels{
		"type": "requests",
	}, AzureDevopsClient.GetRequestCount())

	callback <- func() {
		statsMetrics.GaugeSet(m.prometheus.stats)
	}
}


