package servicequotas

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
)

var (
	lambdaConcurrentExecutionsLimitName = "lambda_concurrent_executions_limit"
	lambdaConcurrentExecutionsLimitDesc = "Measures the maximum number of concurrent executions allowed for an AWS Lambda function."

	lambdaCodeSizeUnzippedLimitBytesName = "lambda_code_size_unzipped_limit_bytes"
	lambdaCodeSizeUnzippedLimitBytesDesc = "Measures the maximum size limit (in bytes) for the unzipped AWS Lambda function code."
)

// LambdaConcurrentExecutionsLimitCheck implements the UsageCheck interface
// for limits for lambda functions
type LambdaConcurrentExecutionsLimitCheck struct {
	client lambdaiface.LambdaAPI
}

// Usage returns the usage and quouta for the lambda concurrent executions and
// lambda code size unzipped limits
func (c *LambdaConcurrentExecutionsLimitCheck) Usage() ([]QuotaUsage, error) {
	param := &lambda.GetAccountSettingsInput{}
	var usages []QuotaUsage

	output, err := c.client.GetAccountSettings(param)
	if err != nil {
		return usages, err
	}

	usages = []QuotaUsage{
		{
			Name:        lambdaConcurrentExecutionsLimitName,
			Description: lambdaConcurrentExecutionsLimitDesc,
			Quota:       float64(*output.AccountLimit.ConcurrentExecutions),
			Usage:       float64(*output.AccountUsage.FunctionCount),
		},
		{
			Name:        lambdaCodeSizeUnzippedLimitBytesName,
			Description: lambdaCodeSizeUnzippedLimitBytesDesc,
			Quota:       float64(*output.AccountLimit.CodeSizeUnzipped),
			Usage:       float64(*output.AccountUsage.TotalCodeSize),
		},
	}

	return usages, err
}
