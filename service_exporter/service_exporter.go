package service_exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// ServiceQuotasExporter AWS service quotas and usage prometheus
// exporter
type ServiceQuotasExporter struct {
	metricsRegion     string
	metricsUsedQuota  map[string]*prometheus.Desc
	metricsQuotaLimit map[string]*prometheus.Desc
}

// NewServiceQuotasExporter creates a new ServiceQuotasExporter
func NewServiceQuotasExporter(region string) *ServiceQuotasExporter {
	return &ServiceQuotasExporter{
		metricsRegion:     region,
		metricsUsedQuota:  map[string]*prometheus.Desc{},
		metricsQuotaLimit: map[string]*prometheus.Desc{},
	}
}

// Describe ...
func (e *ServiceQuotasExporter) Describe(ch chan<- *prometheus.Desc) {
}

// Collect ...
func (e *ServiceQuotasExporter) Collect(ch chan<- prometheus.Metric) {
}
