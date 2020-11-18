package servicequotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/pkg/errors"
)

// Errors returned from this package
var (
	ErrInvalidRegion    = errors.New("invalid region")
	ErrFailedToGetUsage = errors.New("failed to get usage")
)

// QuotaUsageCheck is an interface for retrieving usage for a service quota
type QuotaUsageCheck interface {
	// Code returns the service quota code, eg. "L-0EA8095F"
	Code() string
	// Name returns the name of the servie quota, eg. "Inbound or
	// outbound rules per security group"
	Name() string
	// ServiceName returns the name of the service as reported by
	// the service quotas, eg. "vpc"
	ServiceName() string
	// Usage returns a map of the usage per resource, for example
	// the usage per security group ID ("i-secgrpid": xx) or in case
	// were there is a single resource like for VPCs per region it
	// returns a map with a single entry ("name": xx)
	Usage(awsSession *session.Session, cfgs ...*aws.Config) (map[string]float64, error)
}

// QuotaUsage represents service quota usage
type QuotaUsage struct {
	// Name is the name of the resource, eg. the ID of the VPC or
	// the same as the description for single-resource quotas
	// (eg. VPCs per region)
	Name string
	// Description is the name of the service quota (eg. "Inbound
	// or outbound rules per security group")
	Description string
	// Usage is the current service quota usage
	Usage float64
	// Quota is the current quota
	Quota float64
}

// ServiceQuotas is an implementation for retrieving service quotas
// and their limits using the `QuotaUsageCheck`s
type ServiceQuotas struct {
	session       *session.Session
	region        string
	quotasService *servicequotas.ServiceQuotas
	// maps quota codes to quota limits
	usageChecks map[string]QuotaUsageCheck
}

// QuotasInterface is an interface for retrieving AWS service
// quotas and usage
type QuotasInterface interface {
	QuotasAndUsage() ([]QuotaUsage, error)
}

// NewServiceQuotas creates a ServiceQuotas for `region` and `profile`
// with `usageChecks` or returns an error. Note that the ServiceQuotas
// will only return usage and quotas for the usageChecks passed here
func NewServiceQuotas(region, profile string, usageChecks ...QuotaUsageCheck) (QuotasInterface, error) {
	if !isValidRegion(region) {
		return nil, errors.Wrapf(ErrInvalidRegion, "failed to create ServiceQuotas: %w")
	}

	opts := session.Options{}
	if profile != "" {
		opts = session.Options{Profile: profile}
	}

	awsSession, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	codeToUsageChecks := map[string]QuotaUsageCheck{}
	for _, usageCheck := range usageChecks {
		codeToUsageChecks[usageCheck.Code()] = usageCheck
	}

	quotasService := servicequotas.New(awsSession, aws.NewConfig().WithRegion(region))

	quotas := &ServiceQuotas{
		session:       awsSession,
		region:        region,
		quotasService: quotasService,
		usageChecks:   codeToUsageChecks,
	}
	return quotas, nil
}

func isValidRegion(region string) bool {
	availableRegions := endpoints.AwsPartition().Regions()
	_, ok := availableRegions[region]
	return ok
}

// QuotasAndUsage returns a slice of `QuotaUsage` or an error
func (s *ServiceQuotas) QuotasAndUsage() ([]QuotaUsage, error) {
	quotaUsages := []QuotaUsage{}

	var usageErr error

	params := &servicequotas.ListServiceQuotasInput{ServiceCode: aws.String("vpc")}
	s.quotasService.ListServiceQuotasPages(params,
		func(page *servicequotas.ListServiceQuotasOutput, lastPage bool) bool {
			for _, quota := range page.Quotas {
				if check, ok := s.usageChecks[*quota.QuotaCode]; ok {
					usage, err := check.Usage(s.session, aws.NewConfig().WithRegion(s.region))
					if err != nil {
						usageErr = err
						// stop paging when an error is encountered
						return true
					}

					for name, usage := range usage {
						quotaUsage := QuotaUsage{
							Name:        name,
							Description: check.Name(),
							Usage:       usage,
							Quota:       *quota.Value,
						}
						quotaUsages = append(quotaUsages, quotaUsage)
					}
				}
			}
			return !lastPage
		},
	)

	if usageErr != nil {
		return nil, usageErr
	}

	return quotaUsages, nil
}
