package serviceexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thought-machine/aws-service-quotas-exporter/service_quotas/servicequotas"
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
			{ResourceName: resourceName("i-asdasd1"), Usage: 5, Quota: 10},
			{ResourceName: resourceName("i-asdasd2"), Usage: 2, Quota: 3},
			{ResourceName: resourceName("i-asdasd3"), Usage: 5, Quota: 10},
		},
	}

	exporter := &ServiceQuotasExporter{
		metricsRegion: "eu-west-1",
		quotasClient:  quotasClient,
		metrics: map[string]Metric{
			"i-asdasd1": Metric{usage: 3, limit: 5},
			"i-asdasd2": Metric{usage: 2, limit: 2},
		},
		refreshPeriod: 360,
	}

	exporter.updateMetrics()

	expectedMetrics := map[string]Metric{
		"i-asdasd1": Metric{usage: 5, limit: 10},
		"i-asdasd2": Metric{usage: 2, limit: 3},
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
	}
	quotasClient := &ServiceQuotasMock{
		quotas: []servicequotas.QuotaUsage{firstQ, secondQ},
	}

	exporter := &ServiceQuotasExporter{
		metricsRegion:  region,
		quotasClient:   quotasClient,
		metrics:        map[string]Metric{},
		refreshPeriod:  360,
	}

	exporter.createQuotasAndDescriptions()

	firstUsageDesc := newDesc(region, firstQ.Name, "used_total", "Used amount of desc1", []string{"resource"})
	firstLimitDesc := newDesc(region, firstQ.Name, "limit_total", "Limit of desc1", []string{"resource"})
	secondUsageDesc := newDesc(region, secondQ.Name, "used_total", "Used amount of desc2", []string{"resource"})
	secondLimitDesc := newDesc(region, secondQ.Name, "limit_total", "Limit of desc2", []string{"resource"})
	expectedMetrics := map[string]Metric{
		"i-asdasd1": Metric{
			usageDesc: firstUsageDesc,
			limitDesc: firstLimitDesc,
			usage:     5,
			limit:     10,
		},
		"i-asdasd2": Metric{
			usageDesc: secondUsageDesc,
			limitDesc: secondLimitDesc,
			usage:     1,
			limit:     8,
		},
	}

	assert.Equal(t, expectedMetrics, exporter.metrics)
}
