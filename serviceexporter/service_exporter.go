package serviceexporter

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	logging "github.com/sirupsen/logrus"
	"github.com/thought-machine/aws-service-quotas-exporter/servicequotas"
)

var log = logging.WithFields(logging.Fields{})

// Metric holds usage and limit desc and values
type Metric struct {
	usageDesc   *prometheus.Desc
	limitDesc   *prometheus.Desc
	usage       float64
	limit       float64
	labelValues []string
}

func metricKey(quota servicequotas.QuotaUsage) string {
	return fmt.Sprintf("%s%s", quota.Name, quota.Identifier())
}

// ServiceQuotasExporter AWS service quotas and usage prometheus
// exporter
type ServiceQuotasExporter struct {
	metricsRegion   string
	quotasClient    servicequotas.QuotasInterface
	metrics         map[string]Metric
	metricsLock     *sync.Mutex
	refreshPeriod   int
	waitForMetrics  chan struct{}
	includedAWSTags []string
}

// NewServiceQuotasExporter creates a new ServiceQuotasExporter
func NewServiceQuotasExporter(region, profile string, refreshPeriod int, includedAWSTags []string) (*ServiceQuotasExporter, error) {
	quotasClient, err := servicequotas.NewServiceQuotas(region, profile)
	if err != nil {
		return nil, errors.Wrapf(err, "%w")
	}

	ch := make(chan struct{})
	exporter := &ServiceQuotasExporter{
		metricsRegion:   region,
		quotasClient:    quotasClient,
		metrics:         map[string]Metric{},
		metricsLock:     &sync.Mutex{},
		refreshPeriod:   refreshPeriod,
		waitForMetrics:  ch,
		includedAWSTags: includedAWSTags,
	}
	go exporter.createOrUpdateQuotasAndDescriptions(false)
	go exporter.refreshMetrics()

	return exporter, nil
}

func (e *ServiceQuotasExporter) refreshMetrics() {
	<-e.waitForMetrics

	for {
		time.Sleep(time.Duration(e.refreshPeriod) * time.Second)
		e.createOrUpdateQuotasAndDescriptions(true)
	}
}

func (e *ServiceQuotasExporter) createOrUpdateQuotasAndDescriptions(update bool) {
	quotas, err := e.quotasClient.QuotasAndUsage()
	if err != nil {
		log.Fatalf("Could not retrieve quotas and limits: %s", err)
	}

	e.metricsLock.Lock()
	defer e.metricsLock.Unlock()

	for _, quota := range quotas {
		key := metricKey(quota)
		resourceID := quota.Identifier()

		labels := []string{"resource"}
		labelValues := []string{resourceID}

		for _, tag := range e.includedAWSTags {
			prometheusFormatTag := servicequotas.ToPrometheusNamingFormat(tag)
			labels = append(labels, prometheusFormatTag)
			// Need to set empty label value to keep label name and value count the same
			labelValues = append(labelValues, quota.Tags[prometheusFormatTag])
		}

		if update {
			if resourceMetric, ok := e.metrics[key]; ok {
				log.Infof("Updating metrics for resource (%s)", resourceID)
				resourceMetric.usage = quota.Usage
				resourceMetric.limit = quota.Quota
				resourceMetric.labelValues = labelValues
				e.metrics[key] = resourceMetric
			}
		} else {
			usageHelp := fmt.Sprintf("Used amount of %s", quota.Description)
			usageDesc := newDesc(e.metricsRegion, quota.Name, "used_total", usageHelp, labels)

			limitHelp := fmt.Sprintf("Limit of %s", quota.Description)
			limitDesc := newDesc(e.metricsRegion, quota.Name, "limit_total", limitHelp, labels)
			resourceMetric := Metric{
				usageDesc:   usageDesc,
				limitDesc:   limitDesc,
				usage:       quota.Usage,
				limit:       quota.Quota,
				labelValues: labelValues,
			}
			e.metrics[key] = resourceMetric
		}
	}

	if !update {
		close(e.waitForMetrics)
	}
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
	e.metricsLock.Lock()
	defer e.metricsLock.Unlock()

	for _, metric := range e.metrics {
		ch <- prometheus.MustNewConstMetric(metric.limitDesc, prometheus.GaugeValue, metric.limit, metric.labelValues...)
		ch <- prometheus.MustNewConstMetric(metric.usageDesc, prometheus.GaugeValue, metric.usage, metric.labelValues...)
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
