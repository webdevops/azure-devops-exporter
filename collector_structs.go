package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MetricCollectorRow struct {
	labels prometheus.Labels
	value  float64
}

func NewMetricCollectorList() *MetricCollectorList {
	ret := MetricCollectorList{}
	ret.Init()
	return &ret
}

func (m *MetricCollectorList) Init() {
	m.list = []MetricCollectorRow{}
}

type MetricCollectorList struct {
	list []MetricCollectorRow
}

func (m *MetricCollectorList) Add(labels prometheus.Labels, value float64) {
	m.list = append(m.list, MetricCollectorRow{labels: labels, value: value})
}

func (m *MetricCollectorList) AddInfo(labels prometheus.Labels) {
	m.list = append(m.list, MetricCollectorRow{labels: labels, value: 1})
}

func (m *MetricCollectorList) AddTime(labels prometheus.Labels, value time.Time) {
	timeValue := timeToFloat64(value)

	if timeValue > 0 {
		m.list = append(m.list, MetricCollectorRow{labels: labels, value: timeValue})
	}
}

func (m *MetricCollectorList) AddDuration(labels prometheus.Labels, value time.Duration) {
	m.list = append(m.list, MetricCollectorRow{labels: labels, value: value.Seconds()})
}

func (m *MetricCollectorList) AddIfNotNil(labels prometheus.Labels, value *float64) {
	if value != nil {
		m.list = append(m.list, MetricCollectorRow{labels: labels, value: *value})
	}
}

func (m *MetricCollectorList) AddIfNotZero(labels prometheus.Labels, value float64) {
	if value != 0 {
		m.list = append(m.list, MetricCollectorRow{labels: labels, value: value})
	}
}

func (m *MetricCollectorList) AddIfGreaterZero(labels prometheus.Labels, value float64) {
	if value > 0 {
		m.list = append(m.list, MetricCollectorRow{labels: labels, value: value})
	}
}

func (m *MetricCollectorList) AddBool(labels prometheus.Labels, state bool) {
	value := float64(0)
	if state {
		value = 1
	}

	m.list = append(m.list, MetricCollectorRow{labels: labels, value: value})
}


func (m *MetricCollectorList) GaugeSet(gauge *prometheus.GaugeVec) {
	for _, metric := range m.list {
		gauge.With(metric.labels).Set(metric.value)
	}
}

func (m *MetricCollectorList) CounterAdd(counter *prometheus.CounterVec) {
	for _, metric := range m.list {
		counter.With(metric.labels).Add(metric.value)
	}
}

func (m *MetricCollectorList) SummarySet(counter *prometheus.SummaryVec) {
	for _, metric := range m.list {
		counter.With(metric.labels).Observe(metric.value)
	}
}
