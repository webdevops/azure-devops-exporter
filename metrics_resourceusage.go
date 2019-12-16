package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorResourceUsage struct {
	CollectorProcessorGeneral

	prometheus struct {
		resourceUsageBuild   *prometheus.GaugeVec
		resourceUsageLicense *prometheus.GaugeVec
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


	m.prometheus.resourceUsageLicense = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_resourceusage_license",
			Help: "Azure DevOps resource usage for license informations",
		},
		[]string{
			"name",
		},
	)

	prometheus.MustRegister(m.prometheus.resourceUsageBuild)
	prometheus.MustRegister(m.prometheus.resourceUsageLicense)
}

func (m *MetricsCollectorResourceUsage) Reset() {
	m.prometheus.resourceUsageBuild.Reset()
	m.prometheus.resourceUsageLicense.Reset()
}

func (m *MetricsCollectorResourceUsage) Collect(ctx context.Context, callback chan<- func()) {
	m.CollectResourceUsageBuild(ctx, callback)
	m.CollectResourceUsageAgent(ctx, callback)
}

func (m *MetricsCollectorResourceUsage) CollectResourceUsageAgent(ctx context.Context, callback chan<- func()) {
	resourceUsage, err := AzureDevopsClient.GetResourceUsageAgent()
	if err != nil {
		Logger.Errorf("call[GetResourceUsageAgent]: %v", err)
		return
	}

	resourceUsageMetric := NewMetricCollectorList()

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

	callback <- func() {
		resourceUsageMetric.GaugeSet(m.prometheus.resourceUsageLicense)
	}
}

func (m *MetricsCollectorResourceUsage) CollectResourceUsageBuild(ctx context.Context, callback chan<- func()) {
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
