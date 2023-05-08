package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"github.com/webdevops/go-common/utils/to"
	"go.uber.org/zap"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorAgentPool struct {
	collector.Processor

	prometheus struct {
		agentPool            *prometheus.GaugeVec
		agentPoolSize        *prometheus.GaugeVec
		agentPoolUsage       *prometheus.GaugeVec
		agentPoolAgent       *prometheus.GaugeVec
		agentPoolAgentStatus *prometheus.GaugeVec
		agentPoolAgentJob    *prometheus.GaugeVec
		agentPoolQueueLength *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorAgentPool) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

	m.prometheus.agentPool = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_info",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolID",
			"agentPoolName",
			"agentPoolType",
			"isHosted",
		},
	)
	m.Collector.RegisterMetricList("agentPool", m.prometheus.agentPool, true)

	m.prometheus.agentPoolSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_size",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolID",
		},
	)
	m.Collector.RegisterMetricList("agentPoolSize", m.prometheus.agentPoolSize, true)

	m.prometheus.agentPoolUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_usage",
			Help: "Azure DevOps agentpool usage",
		},
		[]string{
			"agentPoolID",
		},
	)
	m.Collector.RegisterMetricList("agentPoolUsage", m.prometheus.agentPoolUsage, true)

	m.prometheus.agentPoolAgent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_agent_info",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolID",
			"agentPoolAgentID",
			"agentPoolAgentName",
			"agentPoolAgentVersion",
			"provisioningState",
			"maxParallelism",
			"agentPoolAgentOs",
			"enabled",
			"status",
			"hasAssignedRequest",
		},
	)
	m.Collector.RegisterMetricList("agentPoolAgent", m.prometheus.agentPoolAgent, true)

	m.prometheus.agentPoolAgentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_agent_status",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolAgentID",
			"type",
		},
	)
	m.Collector.RegisterMetricList("agentPoolAgentStatus", m.prometheus.agentPoolAgentStatus, true)

	m.prometheus.agentPoolAgentJob = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_agent_job",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolAgentID",
			"jobRequestId",
			"definitionID",
			"definitionName",
			"planType",
			"scopeID",
		},
	)
	m.Collector.RegisterMetricList("agentPoolAgentJob", m.prometheus.agentPoolAgentJob, true)

	m.prometheus.agentPoolQueueLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_queue_length",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolID",
		},
	)
	m.Collector.RegisterMetricList("agentPoolQueueLength", m.prometheus.agentPoolQueueLength, true)
}

func (m *MetricsCollectorAgentPool) Reset() {}

func (m *MetricsCollectorAgentPool) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))
		m.collectAgentInfo(ctx, projectLogger, callback, project)
	}

	for _, agentPoolId := range AzureDevopsServiceDiscovery.AgentPoolList() {
		agentPoolLogger := logger.With(zap.Int64("agentPoolId", agentPoolId))
		m.collectAgentQueues(ctx, agentPoolLogger, callback, agentPoolId)
		m.collectAgentPoolJobs(ctx, agentPoolLogger, callback, agentPoolId)
	}
}

func (m *MetricsCollectorAgentPool) collectAgentInfo(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListAgentQueues(project.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	agentPoolInfoMetric := m.Collector.GetMetricList("agentPool")
	agentPoolSizeMetric := m.Collector.GetMetricList("agentPoolSize")

	for _, agentQueue := range list.List {
		agentPoolInfoMetric.Add(prometheus.Labels{
			"agentPoolID":   int64ToString(agentQueue.Pool.Id),
			"agentPoolName": agentQueue.Name,
			"isHosted":      to.BoolString(agentQueue.Pool.IsHosted),
			"agentPoolType": agentQueue.Pool.PoolType,
		}, 1)

		agentPoolSizeMetric.Add(prometheus.Labels{
			"agentPoolID": int64ToString(agentQueue.Pool.Id),
		}, float64(agentQueue.Pool.Size))
	}
}

func (m *MetricsCollectorAgentPool) collectAgentQueues(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), agentPoolId int64) {
	list, err := AzureDevopsClient.ListAgentPoolAgents(agentPoolId)
	if err != nil {
		logger.Error(err)
		return
	}

	agentPoolUsageMetric := m.Collector.GetMetricList("agentPoolUsage")
	agentPoolAgentMetric := m.Collector.GetMetricList("agentPoolAgent")
	agentPoolAgentStatusMetric := m.Collector.GetMetricList("agentPoolAgentStatus")
	agentPoolAgentJobMetric := m.Collector.GetMetricList("agentPoolAgentJob")

	agentPoolSize := 0
	agentPoolUsed := 0
	for _, agentPoolAgent := range list.List {
		agentPoolSize++
		infoLabels := prometheus.Labels{
			"agentPoolID":           int64ToString(agentPoolId),
			"agentPoolAgentID":      int64ToString(agentPoolAgent.Id),
			"agentPoolAgentName":    agentPoolAgent.Name,
			"agentPoolAgentVersion": agentPoolAgent.Version,
			"provisioningState":     agentPoolAgent.ProvisioningState,
			"maxParallelism":        int64ToString(agentPoolAgent.MaxParallelism),
			"agentPoolAgentOs":      agentPoolAgent.OsDescription,
			"enabled":               to.BoolString(agentPoolAgent.Enabled),
			"status":                agentPoolAgent.Status,
			"hasAssignedRequest":    to.BoolString(agentPoolAgent.AssignedRequest.RequestId > 0),
		}

		agentPoolAgentMetric.Add(infoLabels, 1)

		statusCreatedLabels := prometheus.Labels{
			"agentPoolAgentID": int64ToString(agentPoolAgent.Id),
			"type":             "created",
		}
		agentPoolAgentStatusMetric.Add(statusCreatedLabels, timeToFloat64(agentPoolAgent.CreatedOn))

		if agentPoolAgent.AssignedRequest.RequestId > 0 {
			agentPoolUsed++
			jobLabels := prometheus.Labels{
				"agentPoolAgentID": int64ToString(agentPoolAgent.Id),
				"planType":         agentPoolAgent.AssignedRequest.PlanType,
				"jobRequestId":     int64ToString(agentPoolAgent.AssignedRequest.RequestId),
				"definitionID":     int64ToString(agentPoolAgent.AssignedRequest.Definition.Id),
				"definitionName":   agentPoolAgent.AssignedRequest.Definition.Name,
				"scopeID":          agentPoolAgent.AssignedRequest.ScopeId,
			}
			agentPoolAgentJobMetric.Add(jobLabels, timeToFloat64(*agentPoolAgent.AssignedRequest.AssignTime))
		}
	}

	usage := float64(0)
	if agentPoolSize > 0 {
		usage = float64(agentPoolUsed) / float64(agentPoolSize)
	}
	agentPoolUsageMetric.Add(prometheus.Labels{
		"agentPoolID": int64ToString(agentPoolId),
	}, usage)
}

func (m *MetricsCollectorAgentPool) collectAgentPoolJobs(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), agentPoolId int64) {
	list, err := AzureDevopsClient.ListAgentPoolJobs(agentPoolId)
	if err != nil {
		logger.Error(err)
		return
	}

	agentPoolQueueLengthMetric := m.Collector.GetMetricList("agentPoolQueueLength")

	notStartedJobCount := 0

	for _, agentPoolJob := range list.List {
		if agentPoolJob.AssignTime == nil {
			notStartedJobCount++
		}
	}

	infoLabels := prometheus.Labels{
		"agentPoolID": int64ToString(agentPoolId),
	}

	agentPoolQueueLengthMetric.Add(infoLabels, float64(notStartedJobCount))
}
