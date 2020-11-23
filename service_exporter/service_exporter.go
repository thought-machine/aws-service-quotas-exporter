package serviceexporter

import (
	"fmt"

	"service_quotas/servicequotas"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	logging "github.com/sirupsen/logrus"
)

var log = logging.WithFields(logging.Fields{})

// ServiceQuotasExporter AWS service quotas and usage prometheus
// exporter
type ServiceQuotasExporter struct {
	metricsRegion     string
	quotasClient      servicequotas.QuotasInterface
	metricsUsedQuota  map[string]*prometheus.Desc
	metricsQuotaLimit map[string]*prometheus.Desc
}

// NewServiceQuotasExporter creates a new ServiceQuotasExporter
func NewServiceQuotasExporter(region, profile string) (*ServiceQuotasExporter, error) {
	quotasClient, err := servicequotas.NewServiceQuotas(region, profile)
	if err != nil {
		return nil, errors.Wrapf(err, "%w")
	}

	exporter := &ServiceQuotasExporter{
		metricsRegion:     region,
		quotasClient:      quotasClient,
		metricsUsedQuota:  map[string]*prometheus.Desc{},
		metricsQuotaLimit: map[string]*prometheus.Desc{},
	}
	return exporter, nil
}

// Describe writes descriptors to the prometheus desc channel
func (e *ServiceQuotasExporter) Describe(ch chan<- *prometheus.Desc) {
	quotas, err := e.quotasClient.QuotasAndUsage()
	if err != nil {
		log.Errorf("Can not retrieve quotas and limits: %s", err)
	}

	for _, quota := range quotas {
		// check so we don't report the same metric more than once
		if _, ok := e.metricsUsedQuota[quota.Name]; ok {
			continue
		}

		usedDescription := fmt.Sprintf("Used amount of %s", quota.Description)
		limitDescription := fmt.Sprintf("Limit of %s", quota.Description)

		e.metricsUsedQuota[quota.Name] = newServerMetric(
			e.metricsRegion,
			quota.ServiceName,
			"used_total",
			usedDescription,
			[]string{"resource"},
		)
		e.metricsQuotaLimit[quota.Name] = newServerMetric(
			e.metricsRegion,
			quota.ServiceName,
			"limit_total",
			limitDescription,
			[]string{"resource"},
		)

		ch <- e.metricsUsedQuota[quota.Name]
		ch <- e.metricsQuotaLimit[quota.Name]
	}
}

// Collect implements the collect function for prometheus collectors
func (e *ServiceQuotasExporter) Collect(ch chan<- prometheus.Metric) {
	quotas, err := e.quotasClient.QuotasAndUsage()
	if err != nil {
		log.Errorf("Can not retrieve quotas and limits: %s", err)
	}

	for _, quota := range quotas {
		// validate if the metric is described
		metricLimit, ok := e.metricsQuotaLimit[quota.Name]
		if !ok {
			log.Errorf("metric %s not described", quota.Name)
			continue
		}

		ch <- prometheus.MustNewConstMetric(metricLimit, prometheus.GaugeValue, quota.Quota, quota.Name)

		metricUsed, ok := e.metricsUsedQuota[quota.Name]
		if !ok {
			log.Errorf("metric %s not described", quota.Name)
			continue
		}

		ch <- prometheus.MustNewConstMetric(metricUsed, prometheus.GaugeValue, quota.Usage, quota.Name)
	}
}

func newServerMetric(region, service, metricName, description string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("aws", service, metricName),
		description,
		labels,
		prometheus.Labels{"region": region},
	)
}
