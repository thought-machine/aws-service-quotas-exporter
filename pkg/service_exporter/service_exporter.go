package serviceexporter

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	logging "github.com/sirupsen/logrus"
	"github.com/thought-machine/aws-service-quotas-exporter/pkg/service_quotas/servicequotas"
)

var log = logging.WithFields(logging.Fields{})

// Metric holds usage and limit desc and values
type Metric struct {
	usageDesc *prometheus.Desc
	limitDesc *prometheus.Desc
	usage     float64
	limit     float64
}

// ServiceQuotasExporter AWS service quotas and usage prometheus
// exporter
type ServiceQuotasExporter struct {
	metricsRegion  string
	quotasClient   servicequotas.QuotasInterface
	metrics        map[string]Metric
	refreshPeriod  int
	waitForMetrics chan struct{}
}

// NewServiceQuotasExporter creates a new ServiceQuotasExporter
func NewServiceQuotasExporter(region, profile string, refreshPeriod int) (*ServiceQuotasExporter, error) {
	quotasClient, err := servicequotas.NewServiceQuotas(region, profile)
	if err != nil {
		return nil, errors.Wrapf(err, "%w")
	}

	ch := make(chan struct{})
	exporter := &ServiceQuotasExporter{
		metricsRegion:  region,
		quotasClient:   quotasClient,
		metrics:        map[string]Metric{},
		refreshPeriod:  refreshPeriod,
		waitForMetrics: ch,
	}
	go exporter.createQuotasAndDescriptions()
	go exporter.refreshMetrics()

	return exporter, nil
}

func (e *ServiceQuotasExporter) refreshMetrics() {
	<-e.waitForMetrics

	for {
		time.Sleep(time.Duration(e.refreshPeriod) * time.Second)
		e.updateMetrics()
	}
}

func (e *ServiceQuotasExporter) updateMetrics() {
	quotas, err := e.quotasClient.QuotasAndUsage()
	if err != nil {
		log.Fatalf("Could not retrieve quotas and limits: %s", err)
	}

	for _, quota := range quotas {
		resourceID := quota.Identifier()
		log.Infof("Refreshing metrics for resource (%s)", resourceID)
		if resourceMetric, ok := e.metrics[resourceID]; ok {
			resourceMetric.usage = quota.Usage
			resourceMetric.limit = quota.Quota
			e.metrics[resourceID] = resourceMetric
		}
	}
}

func (e *ServiceQuotasExporter) createQuotasAndDescriptions() {
	quotas, err := e.quotasClient.QuotasAndUsage()
	if err != nil {
		log.Fatalf("Could not retrieve quotas and limits: %s", err)
	}

	for _, quota := range quotas {
		// check so we don't report the same metric more than once
		resourceID := quota.Identifier()
		if _, ok := e.metrics[resourceID]; ok {
			continue
		}

		usageHelp := fmt.Sprintf("Used amount of %s", quota.Description)
		usageDesc := newDesc(e.metricsRegion, quota.Name, "used_total", usageHelp, []string{"resource"})

		limitHelp := fmt.Sprintf("Limit of %s", quota.Description)
		limitDesc := newDesc(e.metricsRegion, quota.Name, "limit_total", limitHelp, []string{"resource"})

		resourceMetric := Metric{
			usageDesc: usageDesc,
			limitDesc: limitDesc,
			usage:     quota.Usage,
			limit:     quota.Quota,
		}
		e.metrics[resourceID] = resourceMetric
	}

	close(e.waitForMetrics)
}

// Describe writes descriptors to the prometheus desc channel
func (e *ServiceQuotasExporter) Describe(ch chan<- *prometheus.Desc) {
	<-e.waitForMetrics

	for _, metric := range e.metrics {
		ch <- metric.usageDesc
		ch <- metric.limitDesc
	}
}

// Collect implements the collect function for prometheus collectors
func (e *ServiceQuotasExporter) Collect(ch chan<- prometheus.Metric) {
	for resourceID, metric := range e.metrics {
		ch <- prometheus.MustNewConstMetric(metric.limitDesc, prometheus.GaugeValue, metric.limit, resourceID)
		ch <- prometheus.MustNewConstMetric(metric.usageDesc, prometheus.GaugeValue, metric.usage, resourceID)
	}
}

func newDesc(region, quotaName, metricName, help string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("aws", quotaName, metricName),
		help,
		labels,
		prometheus.Labels{"region": region},
	)
}
