package main

import (
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorAgentPool struct {
	CollectorProcessorAgentPool

	prometheus struct {
		agentPool            *prometheus.GaugeVec
		agentPoolSize        *prometheus.GaugeVec
		agentPoolAgent       *prometheus.GaugeVec
		agentPoolAgentStatus *prometheus.GaugeVec
		agentPoolAgentJob    *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorAgentPool) Setup(collector *CollectorAgentPool) {
	m.CollectorReference = collector

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

	m.prometheus.agentPoolSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_size",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolID",
		},
	)

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
			"hasAssignedRequest"
		},
	)

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

	prometheus.MustRegister(m.prometheus.agentPool)
	prometheus.MustRegister(m.prometheus.agentPoolSize)
	prometheus.MustRegister(m.prometheus.agentPoolAgent)
	prometheus.MustRegister(m.prometheus.agentPoolAgentStatus)
	prometheus.MustRegister(m.prometheus.agentPoolAgentJob)
}

func (m *MetricsCollectorAgentPool) Reset() {
	m.prometheus.agentPool.Reset()
	m.prometheus.agentPoolSize.Reset()
	m.prometheus.agentPoolAgent.Reset()
	m.prometheus.agentPoolAgentStatus.Reset()
	m.prometheus.agentPoolAgentJob.Reset()
}

func (m *MetricsCollectorAgentPool) Collect(ctx context.Context, callback chan<- func()) {
	for _, project := range m.CollectorReference.azureDevOpsProjects.List {
		m.collectAgentInfo(ctx, callback, project)
	}

	for _, agentPoolId := range m.CollectorReference.AgentPoolIdList {
		m.collectAgentQueues(ctx, callback, agentPoolId)
	}
}

func (m *MetricsCollectorAgentPool) collectAgentInfo(ctx context.Context, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListAgentQueues(project.Id)
	if err != nil {
		Logger.Errorf("agentpool[%v]call[ListAgentQueues]: %v", project.Name, err)
		return
	}

	agentPoolInfoMetric := NewMetricCollectorList()
	agentPoolSizeMetric := NewMetricCollectorList()


	for _, agentQueue := range list.List {
		agentPoolInfoMetric.Add(prometheus.Labels{
			"agentPoolID": int64ToString(agentQueue.Pool.Id),
			"agentPoolName": agentQueue.Name,
			"isHosted": boolToString(agentQueue.Pool.IsHosted),
			"agentPoolType": agentQueue.Pool.PoolType,
		}, 1)

		agentPoolSizeMetric.Add(prometheus.Labels{
			"agentPoolID": int64ToString(agentQueue.Pool.Id),
		},float64(agentQueue.Pool.Size))
	}

	callback <- func() {
		agentPoolInfoMetric.GaugeSet(m.prometheus.agentPool)
		agentPoolSizeMetric.GaugeSet(m.prometheus.agentPoolSize)
	}
}

func (m *MetricsCollectorAgentPool) collectAgentQueues(ctx context.Context, callback chan<- func(), agentPoolId int64) {
	list, err := AzureDevopsClient.ListAgentPoolAgents(agentPoolId)
	if err != nil {
		Logger.Errorf("agentpool[%v]call[ListAgentPoolAgents]: %v", agentPoolId, err)
		return
	}

	agentPoolAgentMetric := NewMetricCollectorList()
	agentPoolAgentStatusMetric := NewMetricCollectorList()
	agentPoolAgentJobMetric := NewMetricCollectorList()

	for _, agentPoolAgent := range list.List {
		infoLabels := prometheus.Labels{
			"agentPoolID":           int64ToString(agentPoolId),
			"agentPoolAgentID":      int64ToString(agentPoolAgent.Id),
			"agentPoolAgentName":    agentPoolAgent.Name,
			"agentPoolAgentVersion": agentPoolAgent.Version,
			"provisioningState":     agentPoolAgent.ProvisioningState,
			"maxParallelism":        int64ToString(agentPoolAgent.MaxParallelism),
			"agentPoolAgentOs":      agentPoolAgent.OsDescription,
			"enabled":               boolToString(agentPoolAgent.Enabled),
			"status":                agentPoolAgent.Status,
			"hasAssignedRequest":    boolToString(agentPoolAgent.AssignedRequest.RequestId > 0)
		}

		agentPoolAgentMetric.Add(infoLabels, 1)

		statusCreatedLabels := prometheus.Labels{
			"agentPoolAgentID": int64ToString(agentPoolAgent.Id),
			"type":             "created",
		}
		agentPoolAgentStatusMetric.Add(statusCreatedLabels, timeToFloat64(agentPoolAgent.CreatedOn))

		if agentPoolAgent.AssignedRequest.RequestId > 0 {
			jobLabels := prometheus.Labels{
				"agentPoolAgentID": int64ToString(agentPoolAgent.Id),
				"planType":         agentPoolAgent.AssignedRequest.PlanType,
				"jobRequestId":     int64ToString(agentPoolAgent.AssignedRequest.RequestId),
				"definitionID":     int64ToString(agentPoolAgent.AssignedRequest.Definition.Id),
				"definitionName":   agentPoolAgent.AssignedRequest.Definition.Name,
				"scopeID":          agentPoolAgent.AssignedRequest.ScopeId,
			}
			agentPoolAgentJobMetric.Add(jobLabels, timeToFloat64(agentPoolAgent.AssignedRequest.AssignTime))
		}
	}

	callback <- func() {
		agentPoolAgentMetric.GaugeSet(m.prometheus.agentPoolAgent)
		agentPoolAgentStatusMetric.GaugeSet(m.prometheus.agentPoolAgentStatus)
		agentPoolAgentJobMetric.GaugeSet(m.prometheus.agentPoolAgentJob)
	}
}
