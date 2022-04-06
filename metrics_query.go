package main

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	prometheusCommon "github.com/webdevops/go-common/prometheus"
)

type MetricsCollectorQuery struct {
	CollectorProcessorQuery

	prometheus struct {
		workItemCount *prometheus.GaugeVec
		workItemData  *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorQuery) Setup(collector *CollectorQuery) {
	m.CollectorReference = collector

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
	prometheus.MustRegister(m.prometheus.workItemCount)

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
	prometheus.MustRegister(m.prometheus.workItemData)
}

func (m *MetricsCollectorQuery) Reset() {
	m.prometheus.workItemCount.Reset()
}

func (m *MetricsCollectorQuery) Collect(ctx context.Context, logger *log.Entry, callback chan<- func()) {
	for _, query := range m.CollectorReference.QueryList {
		queryPair := strings.Split(query, "@")
		m.collectQueryResults(ctx, logger, callback, queryPair[0], queryPair[1])
	}
}

func (m *MetricsCollectorQuery) collectQueryResults(ctx context.Context, logger *log.Entry, callback chan<- func(), queryPath string, projectID string) {
	workItemsMetric := prometheusCommon.NewMetricsList()
	workItemsDataMetric := prometheusCommon.NewMetricsList()

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

	callback <- func() {
		workItemsMetric.GaugeSet(m.prometheus.workItemCount)
		workItemsDataMetric.GaugeSet(m.prometheus.workItemData)
	}
}
