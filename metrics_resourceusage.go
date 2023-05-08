package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"go.uber.org/zap"
)

type MetricsCollectorResourceUsage struct {
	collector.Processor

	prometheus struct {
		resourceUsageBuild   *prometheus.GaugeVec
		resourceUsageLicense *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorResourceUsage) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

	m.prometheus.resourceUsageBuild = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_resourceusage_build",
			Help: "Azure DevOps resource usage for build",
		},
		[]string{
			"name",
		},
	)
	m.Collector.RegisterMetricList("resourceUsageBuild", m.prometheus.resourceUsageBuild, true)

	m.prometheus.resourceUsageLicense = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_resourceusage_license",
			Help: "Azure DevOps resource usage for license informations",
		},
		[]string{
			"name",
		},
	)
	m.Collector.RegisterMetricList("resourceUsageLicense", m.prometheus.resourceUsageLicense, true)
}

func (m *MetricsCollectorResourceUsage) Reset() {}

func (m *MetricsCollectorResourceUsage) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	m.collectResourceUsageBuild(ctx, logger, callback)
	m.collectResourceUsageAgent(ctx, logger, callback)
}

func (m *MetricsCollectorResourceUsage) collectResourceUsageAgent(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func()) {
	resourceUsage, err := AzureDevopsClient.GetResourceUsageAgent()
	if err != nil {
		logger.Error(err)
		return
	}

	resourceUsageMetric := m.Collector.GetMetricList("resourceUsageLicense")

	licenseDetails := resourceUsage.Data.Provider.TaskHubLicenseDetails

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "FreeLicenseCount",
	}, licenseDetails.FreeLicenseCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "FreeHostedLicenseCount",
	}, licenseDetails.FreeHostedLicenseCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "EnterpriseUsersCount",
	}, licenseDetails.EnterpriseUsersCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "EnterpriseUsersCount",
	}, licenseDetails.EnterpriseUsersCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "PurchasedHostedLicenseCount",
	}, licenseDetails.PurchasedHostedLicenseCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "PurchasedHostedLicenseCount",
	}, licenseDetails.PurchasedHostedLicenseCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "TotalLicenseCount",
	}, licenseDetails.TotalLicenseCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "MsdnUsersCount",
	}, licenseDetails.MsdnUsersCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "HostedAgentMinutesFreeCount",
	}, licenseDetails.HostedAgentMinutesFreeCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "HostedAgentMinutesUsedCount",
	}, licenseDetails.HostedAgentMinutesUsedCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "TotalPrivateLicenseCount",
	}, licenseDetails.TotalPrivateLicenseCount)

	resourceUsageMetric.AddIfNotNil(prometheus.Labels{
		"name": "TotalHostedLicenseCount",
	}, licenseDetails.TotalHostedLicenseCount)
}

func (m *MetricsCollectorResourceUsage) collectResourceUsageBuild(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func()) {
	resourceUsage, err := AzureDevopsClient.GetResourceUsageBuild()
	if err != nil {
		logger.Error(err)
		return
	}

	resourceUsageMetric := m.Collector.GetMetricList("resourceUsageBuild")

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
}
