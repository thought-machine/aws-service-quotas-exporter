package servicequotas

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/stretchr/testify/assert"
)

type mockedLambdaClient struct {
	lambdaiface.LambdaAPI
	Output *lambda.GetAccountSettingsOutput
	Err    error
}

func (m *mockedLambdaClient) GetAccountSettings(input *lambda.GetAccountSettingsInput) (*lambda.GetAccountSettingsOutput, error) {
	return m.Output, m.Err
}

func TestLambdaConcurrentExecutionsLimitCheck_Usage(t *testing.T) {
	tests := []struct {
		name          string
		client        lambdaiface.LambdaAPI
		output        *lambda.GetAccountSettingsOutput
		err           error
		expectedUsage []QuotaUsage
		expectedErr   error
	}{
		{
			name: "success",
			client: &mockedLambdaClient{
				Output: &lambda.GetAccountSettingsOutput{
					AccountLimit: &lambda.AccountLimit{
						ConcurrentExecutions: int64p(100),
						CodeSizeUnzipped:     int64p(1000000),
						CodeSizeZipped:       int64p(500000),
					},
					AccountUsage: &lambda.AccountUsage{
						FunctionCount: int64p(50),
						TotalCodeSize: int64p(500000),
					},
				},
				Err: nil,
			},
			expectedUsage: []QuotaUsage{
				{
					Name:        "lambda_concurrent_executions_limit",
					Description: "Measures the maximum number of concurrent executions allowed for an AWS Lambda function.",
					Quota:       100,
					Usage:       50,
				},
				{
					Name:        "lambda_code_size_unzipped_limit_bytes",
					Description: "Measures the maximum size limit (in bytes) for the unzipped AWS Lambda function code.",
					Quota:       1000000,
					Usage:       500000,
				},
			},
			expectedErr: nil,
		},
		{
			name: "error",
			client: &mockedLambdaClient{
				Output: nil,
				Err:    errors.New("some error occurred"),
			},
			expectedUsage: nil,
			expectedErr:   errors.New("some error occurred"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &LambdaConcurrentExecutionsLimitCheck{
				client: test.client,
			}

			usages, err := c.Usage()

			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedUsage, usages)
		})
	}
}

func int64p(i int64) *int64 {
	return &i
}
