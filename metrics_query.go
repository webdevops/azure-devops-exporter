package main

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/prometheus/collector"
	"go.uber.org/zap"
)

type MetricsCollectorQuery struct {
	collector.Processor

	prometheus struct {
		workItemCount *prometheus.GaugeVec
		workItemData  *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorQuery) Setup(collector *collector.Collector) {
	m.Processor.Setup(collector)

	m.prometheus.workItemCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_query_result",
			Help: "Azure DevOps Query Result",
		},
		[]string{
			// We use this only for bugs. Add more fields as needed.
			"projectId",
			"queryPath",
		},
	)
	m.Collector.RegisterMetricList("workItemCount", m.prometheus.workItemCount, true)

	m.prometheus.workItemData = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_workitem_data",
			Help: "Azure DevOps WorkItems",
		},
		[]string{
			"projectId",
			"queryPath",
			"id",
			"title",
			"path",
			"createdDate",
			"acceptedDate",
			"resolvedDate",
			"closedDate",
		},
	)
	m.Collector.RegisterMetricList("workItemData", m.prometheus.workItemData, true)
}

func (m *MetricsCollectorQuery) Reset() {}

func (m *MetricsCollectorQuery) Collect(callback chan<- func()) {
	ctx := m.Context()
	logger := m.Logger()

	for _, project := range AzureDevopsServiceDiscovery.ProjectList() {
		projectLogger := logger.With(zap.String("project", project.Name))

		for _, query := range Opts.AzureDevops.QueriesWithProjects {
			queryPair := strings.Split(query, "@")
			m.collectQueryResults(ctx, projectLogger, callback, queryPair[0], queryPair[1])
		}
	}
}

func (m *MetricsCollectorQuery) collectQueryResults(ctx context.Context, logger *zap.SugaredLogger, callback chan<- func(), queryPath string, projectID string) {
	workItemsMetric := m.Collector.GetMetricList("workItemCount")
	workItemsDataMetric := m.Collector.GetMetricList("workItemData")

	workItemInfoList, err := AzureDevopsClient.QueryWorkItems(queryPath, projectID)
	if err != nil {
		logger.Error(err)
		return
	}

	workItemsMetric.Add(prometheus.Labels{
		"projectId": projectID,
		"queryPath": queryPath,
	}, float64(len(workItemInfoList.List)))

	for _, workItemInfo := range workItemInfoList.List {
		workItem, err := AzureDevopsClient.GetWorkItem(workItemInfo.Url)
		if err != nil {
			logger.Error(err)
			return
		}

		workItemsDataMetric.AddInfo(prometheus.Labels{
			"projectId":    projectID,
			"queryPath":    queryPath,
			"id":           int64ToString(workItem.Id),
			"title":        workItem.Fields.Title,
			"path":         workItem.Fields.Path,
			"createdDate":  workItem.Fields.CreatedDate,
			"acceptedDate": workItem.Fields.AcceptedDate,
			"resolvedDate": workItem.Fields.ResolvedDate,
			"closedDate":   workItem.Fields.ClosedDate,
		})
	}
}
