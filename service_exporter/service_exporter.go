package serviceexporter

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	logging "github.com/sirupsen/logrus"
	"github.com/thought-machine/aws-service-quotas-exporter/service_quotas/servicequotas"
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
		resourceID := quota.Identifier()
		if _, ok := e.metricsUsedQuota[resourceID]; ok {
			continue
		}

		usedDescription := fmt.Sprintf("Used amount of %s", quota.Description)
		e.metricsUsedQuota[resourceID] = newMetric(
			e.metricsRegion,
			quota.Name,
			"used_total",
			usedDescription,
			[]string{"resource"},
		)

		limitDescription := fmt.Sprintf("Limit of %s", quota.Description)
		e.metricsQuotaLimit[resourceID] = newMetric(
			e.metricsRegion,
			quota.Name,
			"limit_total",
			limitDescription,
			[]string{"resource"},
		)

		ch <- e.metricsUsedQuota[resourceID]
		ch <- e.metricsQuotaLimit[resourceID]
	}
}

// Collect implements the collect function for prometheus collectors
func (e *ServiceQuotasExporter) Collect(ch chan<- prometheus.Metric) {
	quotas, err := e.quotasClient.QuotasAndUsage()
	if err != nil {
		log.Errorf("Can not retrieve quotas and limits: %s", err)
	}

	for _, quota := range quotas {
		resourceID := quota.Identifier()

		metricLimit, ok := e.metricsQuotaLimit[resourceID]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(metricLimit, prometheus.GaugeValue, quota.Quota, resourceID)

		metricUsed, ok := e.metricsUsedQuota[resourceID]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(metricUsed, prometheus.GaugeValue, quota.Usage, resourceID)
	}
}

func newMetric(region, quotaName, metricName, description string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("aws", quotaName, metricName),
		description,
		labels,
		prometheus.Labels{"region": region},
	)
}
