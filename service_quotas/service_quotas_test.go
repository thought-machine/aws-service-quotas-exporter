package servicequotas

import (
	"testing"

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
