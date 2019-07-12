package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"time"
	"crypto/sha1"
)

type MetricCollectorRow struct {
	labels prometheus.Labels
	value  float64
}

type MetricCollectorList struct {
	list map[string]MetricCollectorRow
}

func NewMetricCollectorList() (*MetricCollectorList) {
	ret := MetricCollectorList{}
	ret.Init()
	return &ret
}

func (m *MetricCollectorList) Init() {
	m.list = map[string]MetricCollectorRow{}
}

func (m *MetricCollectorList) hashLabels(labels prometheus.Labels) string {
	list := []string{}

	for key, value := range labels {
		list = append(list, fmt.Sprintf("%s=%s", key, value))
	}

	return fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(list, "#"))))
}

func (m *MetricCollectorList) Add(labels prometheus.Labels, value float64) {
	m.list[m.hashLabels(labels)] = MetricCollectorRow{labels: labels, value: value}
}

func (m *MetricCollectorList) AddInfo(labels prometheus.Labels) {
	m.list[m.hashLabels(labels)] = MetricCollectorRow{labels: labels, value: 1}
}

func (m *MetricCollectorList) AddTime(labels prometheus.Labels, value time.Time) {
	timeValue := timeToFloat64(value)

	if timeValue > 0 {
		m.list[m.hashLabels(labels)] = MetricCollectorRow{labels: labels, value: timeValue}
	}
}

func (m *MetricCollectorList) AddDuration(labels prometheus.Labels, value time.Duration) {
	m.list[m.hashLabels(labels)] = MetricCollectorRow{labels: labels, value: value.Seconds()}
}

func (m *MetricCollectorList) AddIfNotZero(labels prometheus.Labels, value float64) {
	if value != 0 {
		m.list[m.hashLabels(labels)] = MetricCollectorRow{labels: labels, value: value}
	}
}

func (m *MetricCollectorList) AddIfGreaterZero(labels prometheus.Labels, value float64) {
	if value > 0 {
		m.list[m.hashLabels(labels)] = MetricCollectorRow{labels: labels, value: value}
	}
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
