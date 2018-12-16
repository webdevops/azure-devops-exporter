package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MetricCollectorRow struct {
	labels prometheus.Labels
	value float64
}

type MetricCollectorList struct {
	list []MetricCollectorRow
}

func (m *MetricCollectorList) Add(labels prometheus.Labels, value float64) {
	m.list = append(m.list, MetricCollectorRow{labels:labels, value:value})
}

func (m *MetricCollectorList) GaugeSet(gauge *prometheus.GaugeVec) {
	for _, metric := range m.list {
		gauge.With(metric.labels).Set(metric.value)
	}
}
