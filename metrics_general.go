package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	prometheusCommon "github.com/webdevops/go-prometheus-common"
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
		[]string{
			"name",
			"type",
		},
	)

	prometheus.MustRegister(m.prometheus.stats)
}

func (m *MetricsCollectorGeneral) Reset() {
	m.prometheus.stats.Reset()
}

func (m *MetricsCollectorGeneral) Collect(ctx context.Context, logger *log.Entry, callback chan<- func()) {
	m.collectAzureDevopsClientStats(ctx, logger, callback)
	m.collectCollectorStats(ctx, logger, callback)
}

func (m *MetricsCollectorGeneral) collectAzureDevopsClientStats(ctx context.Context, logger *log.Entry, callback chan<- func()) {
	statsMetrics := prometheusCommon.NewMetricsList()

	statsMetrics.Add(prometheus.Labels{
		"name": "dev.azure.com",
		"type": "requests",
	}, AzureDevopsClient.GetRequestCount())

	statsMetrics.Add(prometheus.Labels{
		"name": "dev.azure.com",
		"type": "concurrency",
	}, AzureDevopsClient.GetCurrentConcurrency())

	callback <- func() {
		statsMetrics.GaugeSet(m.prometheus.stats)
	}
}

func (m *MetricsCollectorGeneral) collectCollectorStats(ctx context.Context, logger *log.Entry, callback chan<- func()) {
	statsMetrics := prometheusCommon.NewMetricsList()

	for _, collector := range collectorGeneralList {
		if collector.LastScrapeDuration != nil {
			statsMetrics.AddDuration(prometheus.Labels{
				"name": collector.Name,
				"type": "collectorDuration",
			}, *collector.LastScrapeDuration)
		}
	}

	for _, collector := range collectorAgentPoolList {
		if collector.LastScrapeDuration != nil {
			statsMetrics.AddDuration(prometheus.Labels{
				"name": collector.Name,
				"type": "collectorDuration",
			}, *collector.LastScrapeDuration)
		}
	}

	for _, collector := range collectorProjectList {
		if collector.LastScrapeDuration != nil {
			statsMetrics.AddDuration(prometheus.Labels{
				"name": collector.Name,
				"type": "collectorDuration",
			}, *collector.LastScrapeDuration)
		}
	}

	for _, collector := range collectorQueryList {
		if collector.LastScrapeDuration != nil {
			statsMetrics.AddDuration(prometheus.Labels{
				"name": collector.Name,
				"type": "collectorDuration",
			}, *collector.LastScrapeDuration)
		}
	}

	callback <- func() {
		statsMetrics.GaugeSet(m.prometheus.stats)
	}
}
