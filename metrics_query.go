package main

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorQuery struct {
	CollectorProcessorQuery

	prometheus struct {
		workItemCount *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorQuery) Setup(collector *CollectorQuery) {
	m.CollectorReference = collector

	m.prometheus.workItemCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_query_info",
			Help: "Azure DevOps Query",
		},
		[]string{
			// We use this only for bugs. Add more fields as needed.
			"projectId",
			"queryPath",
		},
	)

	prometheus.MustRegister(m.prometheus.workItemCount)
}

func (m *MetricsCollectorQuery) Reset() {
	m.prometheus.workItemCount.Reset()
}

func (m *MetricsCollectorQuery) Collect(ctx context.Context, callback chan<- func()) {
	for _, query := range m.CollectorReference.QueryList {
		queryPair := strings.Split(query, "@")
		m.collectQueryResults(ctx, callback, queryPair[0], queryPair[1])
	}

}

func (m *MetricsCollectorQuery) collectQueryResults(ctx context.Context, callback chan<- func(), queryPath string, projectId string) {
	workItemsMetric := NewMetricCollectorList()

	workItemList, err := AzureDevopsClient.QueryWorkItems(queryPath, projectId)
	if err != nil {
		Logger.Errorf("Query[%v@%v]call[QueryWorkItems]: %v", queryPath, projectId, err)
		return
	}

	workItemsMetric.Add(prometheus.Labels{
		"projectId": projectId,
		"queryPath": queryPath,
	}, float64(len(workItemList.List)))

	callback <- func() {
		workItemsMetric.GaugeSet(m.prometheus.workItemCount)
	}
}
