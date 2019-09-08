package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorResourceUsage struct {
	CollectorProcessorGeneral

	prometheus struct {
		resourceUsageBuild *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorResourceUsage) Setup(collector *CollectorGeneral) {
	m.CollectorReference = collector

	m.prometheus.resourceUsageBuild = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_resourceusage_build",
			Help: "Azure DevOps resource usage for build",
		},
		[]string{
			"name",
		},
	)

	prometheus.MustRegister(m.prometheus.resourceUsageBuild)
}

func (m *MetricsCollectorResourceUsage) Reset() {
	m.prometheus.resourceUsageBuild.Reset()
}

func (m *MetricsCollectorResourceUsage) Collect(ctx context.Context, callback chan<- func()) {
	m.CollectResourceUsage(ctx, callback)
}

func (m *MetricsCollectorResourceUsage) CollectResourceUsage(ctx context.Context, callback chan<- func()) {
	resourceUsage, err := AzureDevopsClient.GetResourceUsageBuild()
	if err != nil {
		Logger.Errorf("call[GetResourceUsageBuild]: %v", err)
		return
	}

	resourceUsageMetric := NewMetricCollectorList()

	if resourceUsage.DistributedTaskAgents != nil {
		resourceUsageMetric.Add(prometheus.Labels{
			"name": "DistributedTaskAgents",
		}, float64(*resourceUsage.DistributedTaskAgents))
	}

	if resourceUsage.PaidPrivateAgentSlots != nil {
		resourceUsageMetric.Add(prometheus.Labels{
			"name": "PaidPrivateAgentSlots",
		}, float64(*resourceUsage.PaidPrivateAgentSlots))
	}

	if resourceUsage.TotalUsage != nil {
		resourceUsageMetric.Add(prometheus.Labels{
			"name": "TotalUsage",
		}, float64(*resourceUsage.TotalUsage))
	}

	if resourceUsage.XamlControllers != nil {
		resourceUsageMetric.Add(prometheus.Labels{
			"name": "XamlControllers",
		}, float64(*resourceUsage.XamlControllers))
	}

	callback <- func() {
		resourceUsageMetric.GaugeSet(m.prometheus.resourceUsageBuild)
	}

}
