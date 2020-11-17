package service_exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

type ServiceQuotasExporter struct {
	metricsRegion     string
	metricsUsedQuota  map[string]*prometheus.Desc
	metricsQuotaLimit map[string]*prometheus.Desc
}

func NewServiceQuotasExporter(region string) *ServiceQuotasExporter {
	return &ServiceQuotasExporter{
		metricsRegion:     region,
		metricsUsedQuota:  map[string]*prometheus.Desc{},
		metricsQuotaLimit: map[string]*prometheus.Desc{},
	}
}

func (e *ServiceQuotasExporter) Describe(ch chan<- *prometheus.Desc) {
}

func (e *ServiceQuotasExporter) Collect(ch chan<- prometheus.Metric) {
}
