package main

import (
	"sync"
	devopsClient "azure-devops-exporter/src/azure-devops-client"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics run
func runMetricsCollectionAgentPool() {
	var wg sync.WaitGroup

	callbackChannel := make(chan func())

	for _, agentPool := range agentPoolList {
		if len(opts.AzureDevopsFilterAgentPoolId) > 0 {
			if !arrayInt64Contains(opts.AzureDevopsFilterAgentPoolId, agentPool.Pool.Id) {
				continue
			}
		}

		wg.Add(1)
		go func(agentPoolId devopsClient.AgentQueue) {
			defer wg.Done()
			collectAgentPoolAgents(agentPoolId, callbackChannel)
		}(agentPool)
	}

	go func() {
		var callbackList []func()
		for callback := range callbackChannel {
			callbackList = append(callbackList, callback)
		}

		prometheusAgentPoolAgent.Reset()
		prometheusAgentPoolAgentStatus.Reset()
		prometheusAgentPoolAgentJob.Reset()
		for _, callback := range callbackList {
			callback()
		}

		Logger.Messsage("run[queue]: finished")
	}()

	// wait for all funcs
	wg.Wait()
	close(callbackChannel)
}

func collectAgentPoolAgents(agentPool devopsClient.AgentQueue, callback chan<- func()) {
	list, err := AzureDevopsClient.ListAgentPoolAgents(agentPool.Pool.Id)
	if err != nil {
		ErrorLogger.Messsage("agentpool[%v]: %v", agentPool.Pool.Id, err)
		return
	}

	for _, agentPoolAgent := range list.List {
		infoLabels := prometheus.Labels{
			"agentPoolID": int64ToString(agentPool.Pool.Id),
			"agentPoolAgentID": int64ToString(agentPoolAgent.Id),
			"agentPoolAgentName": agentPoolAgent.Name,
			"agentPoolAgentVersion": agentPoolAgent.Version,
			"provisioningState": agentPoolAgent.ProvisioningState,
			"maxParallelism": int64ToString(agentPoolAgent.MaxParallelism),
			"agentPoolAgentOs": agentPoolAgent.OsDescription,
			"enabled": boolToString(agentPoolAgent.Enabled),
		}

		statusCreatedLabels :=prometheus.Labels{
			"agentPoolAgentID": int64ToString(agentPoolAgent.Id),
			"type": "created",
		}
		statusCreatedValue := float64(agentPoolAgent.CreatedOn.Unix())

		callback <- func() {
			prometheusAgentPoolAgent.With(infoLabels).Set(1)
			prometheusAgentPoolAgentStatus.With(statusCreatedLabels).Set(statusCreatedValue)
		}

		if agentPoolAgent.AssignedRequest.RequestId > 0 {
			jobLabels :=prometheus.Labels{
				"agentPoolAgentID": int64ToString(agentPoolAgent.Id),
				"planType": agentPoolAgent.AssignedRequest.PlanType,
				"jobRequestId": int64ToString(agentPoolAgent.AssignedRequest.RequestId),
				"definitionID": int64ToString(agentPoolAgent.AssignedRequest.Definition.Id),
				"definitionName": agentPoolAgent.AssignedRequest.Definition.Name,
				"scopeID": agentPoolAgent.AssignedRequest.ScopeId,
			}
			jobValue := float64(agentPoolAgent.AssignedRequest.AssignTime.Unix())
			callback <- func() {
				prometheusAgentPoolAgentJob.With(jobLabels).Set(jobValue)
			}
		}
	}
}
