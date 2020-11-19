package servicequotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
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

type usageCheck func(c client.ConfigProvider, cfgs ...*aws.Config) ([]QuotaUsage, error)

func newUsageChecks() map[string]usageCheck {
	return map[string]usageCheck{
		"L-0EA8095F": RulesPerSecurityGroupUsage,
		"L-2AFB9258": SecurityGroupsPerENIUsage,
		"L-E79EC296": SecurityGroupsPerRegionUsage,
		"L-34B43A08": StandardSpotInstanceRequestsUsage,
		"L-1216C47A": RunningOnDemandStandardInstancesUsage,
	}
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
// and their limits
type ServiceQuotas struct {
	session       *session.Session
	region        string
	quotasService *servicequotas.ServiceQuotas
}

// QuotasInterface is an interface for retrieving AWS service
// quotas and usage
type QuotasInterface interface {
	QuotasAndUsage() ([]QuotaUsage, error)
}

// NewServiceQuotas creates a ServiceQuotas for `region` and `profile`
// or returns an error. Note that the ServiceQuotas will only return
// usage and quotas for the service quotas with implemented usage checks
func NewServiceQuotas(region, profile string) (QuotasInterface, error) {
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

	quotasService := servicequotas.New(awsSession, aws.NewConfig().WithRegion(region))

	quotas := &ServiceQuotas{
		session:       awsSession,
		region:        region,
		quotasService: quotasService,
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
	allQuotaUsages := []QuotaUsage{}
	usageChecks := newUsageChecks()

	var usageErr error

	params := &servicequotas.ListServiceQuotasInput{}
	s.quotasService.ListServiceQuotasPages(params,
		func(page *servicequotas.ListServiceQuotasOutput, lastPage bool) bool {
			for _, quota := range page.Quotas {
				if check, ok := usageChecks[*quota.QuotaCode]; ok {
					quotaUsages, err := check(s.session, aws.NewConfig().WithRegion(s.region))
					if err != nil {
						usageErr = err
						// stop paging when an error is encountered
						return true
					}

					for _, quotaUsage := range quotaUsages {
						quotaUsage.Quota = *quota.Value
						allQuotaUsages = append(allQuotaUsages, quotaUsage)
					}
				}
			}
			return !lastPage
		},
	)

	if usageErr != nil {
		return nil, usageErr
	}

	return allQuotaUsages, nil
}
