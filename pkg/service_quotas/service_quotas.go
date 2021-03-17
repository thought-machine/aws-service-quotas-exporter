package servicequotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsservicequotas "github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/servicequotas/servicequotasiface"
	"github.com/pkg/errors"
)

// Errors returned from this package
var (
	ErrInvalidRegion       = errors.New("invalid region")
	ErrFailedToListQuotas  = errors.New("failed to list quotas")
	ErrFailedToGetUsage    = errors.New("failed to get usage")
	ErrFailedToConvertCidr = errors.New("failed to convert CIDR block from string to int")
)

func allServices() []string {
	return []string{"ec2", "vpc"}
}

// UsageCheck is an interface for retrieving service quota usage
type UsageCheck interface {
	// Usage returns slice of QuotaUsage or an error
	Usage() ([]QuotaUsage, error)
}

func newUsageChecks(c client.ConfigProvider, cfgs ...*aws.Config) (map[string]UsageCheck, []UsageCheck) {
	// all clients that will be used by the usage checks
	ec2Client := ec2.New(c, cfgs...)
	autoscalingClient := autoscaling.New(c, cfgs...)

	serviceQuotasUsageChecks := map[string]UsageCheck{
		"L-0EA8095F": &RulesPerSecurityGroupUsageCheck{ec2Client},
		"L-2AFB9258": &SecurityGroupsPerENIUsageCheck{ec2Client},
		"L-E79EC296": &SecurityGroupsPerRegionUsageCheck{ec2Client},
		"L-34B43A08": &StandardSpotInstanceRequestsUsageCheck{ec2Client},
		"L-1216C47A": &RunningOnDemandStandardInstancesUsageCheck{ec2Client},
	}

	otherUsageChecks := []UsageCheck{
		&AvailableIpsPerSubnetUsageCheck{ec2Client},
		&ASGUsageCheck{autoscalingClient},
	}

	return serviceQuotasUsageChecks, otherUsageChecks
}

// QuotaUsage represents service quota usage
type QuotaUsage struct {
	// Name is the name of the quota (eg. spot_instance_requests)
	// or the name given to the piece of exported availibility
	// information (eg. available_IPs_per_subnet)
	Name string
	// ResourceName is the name of the resource in case the quota
	// is for multiple resources. As an example for "rules per
	// security group" the ResourceName will be the ARN of the
	// security group.
	ResourceName *string
	// Description is the name of the service quota (eg. "Inbound
	// or outbound rules per security group")
	Description string
	// Usage is the current service quota usage
	Usage float64
	// Quota is the current quota
	Quota float64
}

// Identifier for the service quota. Either the resource name in case
// the quota is for multiple resources or the name of the quota
func (q QuotaUsage) Identifier() string {
	if q.ResourceName != nil {
		return *q.ResourceName
	}
	return q.Name
}

// ServiceQuotas is an implementation for retrieving service quotas
// and their limits
type ServiceQuotas struct {
	session                  *session.Session
	region                   string
	quotasService            servicequotasiface.ServiceQuotasAPI
	serviceQuotasUsageChecks map[string]UsageCheck
	otherUsageChecks         []UsageCheck
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

	quotasService := awsservicequotas.New(awsSession, aws.NewConfig().WithRegion(region))
	serviceQuotasChecks, otherChecks := newUsageChecks(awsSession, aws.NewConfig().WithRegion(region))

	quotas := &ServiceQuotas{
		session:                  awsSession,
		region:                   region,
		quotasService:            quotasService,
		serviceQuotasUsageChecks: serviceQuotasChecks,
		otherUsageChecks:         otherChecks,
	}
	return quotas, nil
}

func isValidRegion(region string) bool {
	availableRegions := endpoints.AwsPartition().Regions()
	_, ok := availableRegions[region]
	return ok
}

func (s *ServiceQuotas) quotasForService(service string) ([]QuotaUsage, error) {
	serviceQuotaUsages := []QuotaUsage{}
	var usageErr error

	params := &awsservicequotas.ListServiceQuotasInput{ServiceCode: aws.String(service)}
	err := s.quotasService.ListServiceQuotasPages(params,
		func(page *awsservicequotas.ListServiceQuotasOutput, lastPage bool) bool {
			if page != nil {
				for _, quota := range page.Quotas {
					if check, ok := s.serviceQuotasUsageChecks[*quota.QuotaCode]; ok {
						quotaUsages, err := check.Usage()
						if err != nil {
							usageErr = err
							// stop paging when an error is encountered
							return true
						}

						for _, quotaUsage := range quotaUsages {
							quotaUsage.Quota = *quota.Value
							serviceQuotaUsages = append(serviceQuotaUsages, quotaUsage)
						}
					}
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToListQuotas, "%w", err)
	}

	if usageErr != nil {
		return nil, usageErr
	}

	return serviceQuotaUsages, nil
}

// QuotasAndUsage returns a slice of `QuotaUsage` or an error
func (s *ServiceQuotas) QuotasAndUsage() ([]QuotaUsage, error) {
	allQuotaUsages := []QuotaUsage{}
	for _, service := range allServices() {
		serviceQuotas, err := s.quotasForService(service)
		if err != nil {
			return nil, err
		}

		for _, quota := range serviceQuotas {
			allQuotaUsages = append(allQuotaUsages, quota)
		}
	}

	for _, check := range s.otherUsageChecks {
		quotas, err := check.Usage()
		if err != nil {
			return nil, err
		}

		for _, quota := range quotas {
			allQuotaUsages = append(allQuotaUsages, quota)
		}
	}

	return allQuotaUsages, nil
}
