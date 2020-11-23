package servicequotas

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awsservicequotas "github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/servicequotas/servicequotasiface"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockServiceQuotasClient struct {
	servicequotasiface.ServiceQuotasAPI

	err                       error
	ListServiceQuotasResponse *awsservicequotas.ListServiceQuotasOutput
}

func (m *mockServiceQuotasClient) ListServiceQuotasPages(input *awsservicequotas.ListServiceQuotasInput, fn func(*awsservicequotas.ListServiceQuotasOutput, bool) bool) error {
	fn(m.ListServiceQuotasResponse, true)
	return m.err
}

type UsageCheckMock struct {
	err    error
	usages []QuotaUsage
}

func (m *UsageCheckMock) Usage() ([]QuotaUsage, error) {
	return m.usages, m.err
}

func TestQuotasAndUsageWithError(t *testing.T) {
	mockClient := &mockServiceQuotasClient{
		err:                       errors.New("some err"),
		ListServiceQuotasResponse: nil,
	}

	serviceQuotas := ServiceQuotas{quotasService: mockClient}
	quotasAndUsage, err := serviceQuotas.QuotasAndUsage()

	assert.Error(t, err)
	assert.Nil(t, quotasAndUsage)
}

func TestQuotasAndUsageWithUsageError(t *testing.T) {
	mockClient := &mockServiceQuotasClient{
		ListServiceQuotasResponse: &awsservicequotas.ListServiceQuotasOutput{
			Quotas: []*awsservicequotas.ServiceQuota{
				{
					QuotaCode: aws.String("L-1234"),
					Value:     aws.Float64(15),
				},
			},
		},
	}

	expectedErr := errors.New("some err")
	usageCheckMock := &UsageCheckMock{
		err:    expectedErr,
		usages: nil,
	}

	serviceQuotas := ServiceQuotas{
		quotasService: mockClient,
		usageChecks: map[string]UsageCheck{
			"L-1234": usageCheckMock,
		},
	}
	quotasAndUsage, err := serviceQuotas.QuotasAndUsage()

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, quotasAndUsage)
}

func TestQuotasAndUsage(t *testing.T) {
	mockClient := &mockServiceQuotasClient{
		ListServiceQuotasResponse: &awsservicequotas.ListServiceQuotasOutput{
			Quotas: []*awsservicequotas.ServiceQuota{
				{
					QuotaCode: aws.String("L-1234"),
					Value:     aws.Float64(15),
				},
				{
					QuotaCode: aws.String("L-NOTIMPLEMENTED"),
					Value:     aws.Float64(6),
				},
				{
					QuotaCode: aws.String("L-5678"),
					Value:     aws.Float64(2),
				},
			},
		},
	}

	firstUsageCheckMock := &UsageCheckMock{
		usages: []QuotaUsage{
			{
				Name:        "i-resource1",
				Description: "check with multiple resources",
				Usage:       10,
			},
			{
				Name:        "i-resource2",
				Description: "check with multiple resources",
				Usage:       3,
			},
		},
	}
	secondUsageCheckMock := &UsageCheckMock{
		usages: []QuotaUsage{
			{
				Name:        "some check",
				Description: "some check",
				Usage:       1,
			},
		},
	}

	serviceQuotas := ServiceQuotas{
		quotasService: mockClient,
		usageChecks: map[string]UsageCheck{
			"L-1234": firstUsageCheckMock,
			"L-5678": secondUsageCheckMock,
		},
	}
	actualQuotasAndUsage, err := serviceQuotas.QuotasAndUsage()

	expectedQuotasAndUsage := []QuotaUsage{
		{
			Name:        "i-resource1",
			Description: "check with multiple resources",
			Usage:       10,
			Quota:       15,
		},
		{
			Name:        "i-resource2",
			Description: "check with multiple resources",
			Usage:       3,
			Quota:       15,
		},
		{
			Name:        "some check",
			Description: "some check",
			Usage:       1,
			Quota:       2,
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedQuotasAndUsage, actualQuotasAndUsage)
}
