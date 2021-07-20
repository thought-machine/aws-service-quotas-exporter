package serviceexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thought-machine/aws-service-quotas-exporter/pkg/service_quotas/servicequotas"
)

func resourceName(name string) *string {
	return &name
}

type ServiceQuotasMock struct {
	quotas []servicequotas.QuotaUsage
	err    error
}

func (s *ServiceQuotasMock) QuotasAndUsage() ([]servicequotas.QuotaUsage, error) {
	return s.quotas, s.err
}

func TestUpdateMetrics(t *testing.T) {
	quotasClient := &ServiceQuotasMock{
		quotas: []servicequotas.QuotaUsage{
			{ResourceName: resourceName("i-asdasd1"), Usage: 5, Quota: 10, Tags: map[string]string{"dummy_tag": "dummy-value"}},
			{ResourceName: resourceName("i-asdasd2"), Usage: 2, Quota: 3},
			{ResourceName: resourceName("i-asdasd3"), Usage: 5, Quota: 10},
		},
	}

	exporter := &ServiceQuotasExporter{
		metricsRegion: "eu-west-1",
		quotasClient:  quotasClient,
		metrics: map[string]Metric{
			"i-asdasd1": Metric{usage: 3, limit: 5, labelValues: []string{"before-dummy-value"}},
			"i-asdasd2": Metric{usage: 2, limit: 2},
		},
		includedAWSTags: []string{"dummy-tag"},
		refreshPeriod:   360,
	}

	exporter.createOrUpdateQuotasAndDescriptions(true)

	expectedMetrics := map[string]Metric{
		"i-asdasd1": Metric{usage: 5, limit: 10, labelValues: []string{"i-asdasd1", "dummy-value"}},
		"i-asdasd2": Metric{usage: 2, limit: 3, labelValues: []string{"i-asdasd2", ""}},
	}
	assert.Equal(t, expectedMetrics, exporter.metrics)
}

func TestCreateQuotasAndDescriptions(t *testing.T) {
	region := "eu-west-1"

	firstQ := servicequotas.QuotaUsage{
		Name:         "Name1",
		ResourceName: resourceName("i-asdasd1"),
		Description:  "desc1",
		Usage:        5,
		Quota:        10,
	}
	secondQ := servicequotas.QuotaUsage{
		Name:         "Name2",
		ResourceName: resourceName("i-asdasd2"),
		Description:  "desc2",
		Usage:        1,
		Quota:        8,
		Tags:         map[string]string{"dummy_tag": "dummy-value", "dummy_tag2": "dummy-value2"},
	}
	quotasClient := &ServiceQuotasMock{
		quotas: []servicequotas.QuotaUsage{firstQ, secondQ},
	}

	ch := make(chan struct{})
	exporter := &ServiceQuotasExporter{
		metricsRegion:   region,
		quotasClient:    quotasClient,
		metrics:         map[string]Metric{},
		refreshPeriod:   360,
		waitForMetrics:  ch,
		includedAWSTags: []string{"dummy-tag", "dummy-tag2"},
	}

	exporter.createOrUpdateQuotasAndDescriptions(false)

	firstUsageDesc := newDesc(region, firstQ.Name, "used_total", "Used amount of desc1", []string{"resource", "dummy_tag", "dummy_tag2"})
	firstLimitDesc := newDesc(region, firstQ.Name, "limit_total", "Limit of desc1", []string{"resource", "dummy_tag", "dummy_tag2"})
	secondUsageDesc := newDesc(region, secondQ.Name, "used_total", "Used amount of desc2", []string{"resource", "dummy_tag", "dummy_tag2"})
	secondLimitDesc := newDesc(region, secondQ.Name, "limit_total", "Limit of desc2", []string{"resource", "dummy_tag", "dummy_tag2"})
	expectedMetrics := map[string]Metric{
		"Name1i-asdasd1": Metric{
			usageDesc:   firstUsageDesc,
			limitDesc:   firstLimitDesc,
			usage:       5,
			limit:       10,
			labelValues: []string{"i-asdasd1", "", ""},
		},
		"Name2i-asdasd2": Metric{
			usageDesc:   secondUsageDesc,
			limitDesc:   secondLimitDesc,
			usage:       1,
			limit:       8,
			labelValues: []string{"i-asdasd2", "dummy-value", "dummy-value2"},
		},
	}

	assert.Equal(t, expectedMetrics, exporter.metrics)
}

func TestCreateQuotasAndDescriptionsRefresh(t *testing.T) {
	quotasClient := &ServiceQuotasMock{
		quotas: []servicequotas.QuotaUsage{
			{ResourceName: resourceName("i-asdasd1"),
				Usage:       5,
				Quota:       10,
				Tags:        map[string]string{"dummy_tag": "dummy-value"},
				Description: "This won't change the metric description for update",
			},
			{ResourceName: resourceName("i-asdasd3"), Usage: 5, Quota: 10},
		},
	}

	desc := newDesc("eu-west-1", "some-quota", "some-metric", "help", []string{})

	ch := make(chan struct{})
	exporter := &ServiceQuotasExporter{
		metricsRegion: "eu-west-1",
		quotasClient:  quotasClient,
		metrics: map[string]Metric{
			"i-asdasd1": Metric{usage: 3, limit: 5, labelValues: []string{"before-dummy-value"}, usageDesc: desc},
		},
		waitForMetrics:  ch,
		includedAWSTags: []string{"dummy-tag"},
		refreshPeriod:   360,
	}

	exporter.createOrUpdateQuotasAndDescriptions(true)

	expectedMetrics := map[string]Metric{
		"i-asdasd1": Metric{usage: 5, limit: 10, labelValues: []string{"i-asdasd1", "dummy-value"}, usageDesc: desc},
	}

	assert.Equal(t, expectedMetrics, exporter.metrics)

	close(ch) // should panic if it was already closed
}
