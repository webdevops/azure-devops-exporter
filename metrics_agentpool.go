package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
	prometheusCommon "github.com/webdevops/go-prometheus-common"
)

type MetricsCollectorAgentPool struct {
	CollectorProcessorAgentPool

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
	prometheus.MustRegister(m.prometheus.agentPool)

	m.prometheus.agentPoolSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_size",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolID",
		},
	)
	prometheus.MustRegister(m.prometheus.agentPoolSize)

	m.prometheus.agentPoolUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_usage",
			Help: "Azure DevOps agentpool usage",
		},
		[]string{
			"agentPoolID",
		},
	)
	prometheus.MustRegister(m.prometheus.agentPoolUsage)

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
	prometheus.MustRegister(m.prometheus.agentPoolAgent)

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
	prometheus.MustRegister(m.prometheus.agentPoolAgentStatus)

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
	prometheus.MustRegister(m.prometheus.agentPoolAgentJob)

	m.prometheus.agentPoolQueueLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_agentpool_queue_length",
			Help: "Azure DevOps agentpool",
		},
		[]string{
			"agentPoolID",
		},
	)
	prometheus.MustRegister(m.prometheus.agentPoolQueueLength)
}

func (m *MetricsCollectorAgentPool) Reset() {
	m.prometheus.agentPool.Reset()
	m.prometheus.agentPoolSize.Reset()
	m.prometheus.agentPoolAgent.Reset()
	m.prometheus.agentPoolAgentStatus.Reset()
	m.prometheus.agentPoolAgentJob.Reset()
	m.prometheus.agentPoolQueueLength.Reset()
}

func (m *MetricsCollectorAgentPool) Collect(ctx context.Context, logger *log.Entry, callback chan<- func()) {
	for _, project := range m.CollectorReference.azureDevOpsProjects.List {
		contextLogger := logger.WithFields(log.Fields{
			"project": project.Name,
		})
		m.collectAgentInfo(ctx, contextLogger, callback, project)
	}

	for _, agentPoolId := range m.CollectorReference.AgentPoolIdList {
		contextLogger := logger.WithFields(log.Fields{
			"agentPoolId": agentPoolId,
		})

		m.collectAgentQueues(ctx, contextLogger, callback, agentPoolId)
		m.collectAgentPoolJobs(ctx, contextLogger, callback, agentPoolId)
	}
}

func (m *MetricsCollectorAgentPool) collectAgentInfo(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListAgentQueues(project.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	agentPoolInfoMetric := prometheusCommon.NewMetricsList()
	agentPoolSizeMetric := prometheusCommon.NewMetricsList()

	for _, agentQueue := range list.List {
		agentPoolInfoMetric.Add(prometheus.Labels{
			"agentPoolID":   int64ToString(agentQueue.Pool.Id),
			"agentPoolName": agentQueue.Name,
			"isHosted":      boolToString(agentQueue.Pool.IsHosted),
			"agentPoolType": agentQueue.Pool.PoolType,
		}, 1)

		agentPoolSizeMetric.Add(prometheus.Labels{
			"agentPoolID": int64ToString(agentQueue.Pool.Id),
		}, float64(agentQueue.Pool.Size))
	}

	callback <- func() {
		agentPoolInfoMetric.GaugeSet(m.prometheus.agentPool)
		agentPoolSizeMetric.GaugeSet(m.prometheus.agentPoolSize)
	}
}

func (m *MetricsCollectorAgentPool) collectAgentQueues(ctx context.Context, logger *log.Entry, callback chan<- func(), agentPoolId int64) {
	list, err := AzureDevopsClient.ListAgentPoolAgents(agentPoolId)
	if err != nil {
		logger.Error(err)
		return
	}

	agentPoolUsageMetric := prometheusCommon.NewMetricsList()
	agentPoolAgentMetric := prometheusCommon.NewMetricsList()
	agentPoolAgentStatusMetric := prometheusCommon.NewMetricsList()
	agentPoolAgentJobMetric := prometheusCommon.NewMetricsList()

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
			"enabled":               boolToString(agentPoolAgent.Enabled),
			"status":                agentPoolAgent.Status,
			"hasAssignedRequest":    boolToString(agentPoolAgent.AssignedRequest.RequestId > 0),
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

	agentPoolUsageMetric.Add(prometheus.Labels{
		"agentPoolID": int64ToString(agentPoolId),
	}, float64(agentPoolUsed)/float64(agentPoolSize))

	callback <- func() {
		agentPoolUsageMetric.GaugeSet(m.prometheus.agentPoolUsage)
		agentPoolAgentMetric.GaugeSet(m.prometheus.agentPoolAgent)
		agentPoolAgentStatusMetric.GaugeSet(m.prometheus.agentPoolAgentStatus)
		agentPoolAgentJobMetric.GaugeSet(m.prometheus.agentPoolAgentJob)
	}
}

func (m *MetricsCollectorAgentPool) collectAgentPoolJobs(ctx context.Context, logger *log.Entry, callback chan<- func(), agentPoolId int64) {
	list, err := AzureDevopsClient.ListAgentPoolJobs(agentPoolId)
	if err != nil {
		logger.Error(err)
		return
	}

	agentPoolQueueLengthMetric := prometheusCommon.NewMetricsList()

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

	callback <- func() {
		agentPoolQueueLengthMetric.GaugeSet(m.prometheus.agentPoolQueueLength)
	}
}
