package servicequotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

// Not all quota limits here are reported under "ec2", but all of the
// usage checks are using the ec2 service
const (
	rulesPerSecGrpName = "rules_per_security_group"
	rulesPerSecGrpDesc = "rules per security group"

	secGroupsPerENIName = "security_groups_per_network_interface"
	secGroupsPerENIDesc = "security groups per network interface"

	securityGroupsPerRegionName = "security_groups_per_region"
	securityGroupsPerRegionDesc = "security groups per region"

	spotInstanceRequestsName = "spot_instance_requests"
	spotInstanceRequestsDesc = "spot instance requests"

	onDemandInstanceRequestsName = "ondemand_instance_requests"
	onDemandInstanceRequestsDesc = "ondemand instance requests"
)

// RulesPerSecurityGroupUsageCheck implements the UsageCheck interface
// for rules per security group
type RulesPerSecurityGroupUsageCheck struct {
	client ec2iface.EC2API
}

// Usage returns the usage for each security group ID with the usage
// value being the sum of their inbound and outbound rules or an error
func (c *RulesPerSecurityGroupUsageCheck) Usage() ([]QuotaUsage, error) {
	quotaUsages := []QuotaUsage{}

	securityGroups := []*ec2.SecurityGroup{}
	params := &ec2.DescribeSecurityGroupsInput{}
	err := c.client.DescribeSecurityGroupsPages(params,
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			if page != nil {
				for _, group := range page.SecurityGroups {
					securityGroups = append(securityGroups, group)
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	for _, securityGroup := range securityGroups {
		inboundRules := len(securityGroup.IpPermissions)
		outboundRules := len(securityGroup.IpPermissionsEgress)
		quotaUsage := QuotaUsage{
			Name:         rulesPerSecGrpName,
			ResourceName: securityGroup.GroupId,
			Description:  rulesPerSecGrpDesc,
			Usage:        float64(inboundRules + outboundRules),
		}
		quotaUsages = append(quotaUsages, quotaUsage)
	}

	return quotaUsages, nil
}

// SecurityGroupsPerENIUsageCheck implements the UsageCheck interface
// for security groups per ENI
type SecurityGroupsPerENIUsageCheck struct {
	client ec2iface.EC2API
}

// Usage returns usage for each Elastic Network Interface ID with the
// usage value being the number of security groups for each ENI or an
// error
func (c *SecurityGroupsPerENIUsageCheck) Usage() ([]QuotaUsage, error) {
	quotaUsages := []QuotaUsage{}

	params := &ec2.DescribeNetworkInterfacesInput{}
	err := c.client.DescribeNetworkInterfacesPages(params,
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			if page != nil {
				for _, eni := range page.NetworkInterfaces {
					usage := QuotaUsage{
						Name:         secGroupsPerENIName,
						ResourceName: eni.NetworkInterfaceId,
						Description:  secGroupsPerENIDesc,
						Usage:        float64(len(eni.Groups)),
					}
					quotaUsages = append(quotaUsages, usage)
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	return quotaUsages, nil
}

// SecurityGroupsPerRegionUsageCheck implements the UsageCheck interface
// for security groups per region
type SecurityGroupsPerRegionUsageCheck struct {
	client ec2iface.EC2API
}

// Usage returns usage for security groups per region as the number of
// all security groups for the region specified with `cfgs` or an error
func (c *SecurityGroupsPerRegionUsageCheck) Usage() ([]QuotaUsage, error) {
	numGroups := 0

	params := &ec2.DescribeSecurityGroupsInput{}
	err := c.client.DescribeSecurityGroupsPages(params,
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			if page != nil {
				numGroups += len(page.SecurityGroups)
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	usage := []QuotaUsage{
		{
			Name:        securityGroupsPerRegionName,
			Description: securityGroupsPerRegionDesc,
			Usage:       float64(numGroups),
		},
	}
	return usage, nil
}

func standardInstanceTypeFilter() *ec2.Filter {
	return &ec2.Filter{
		Name: aws.String("instance-type"),
		Values: []*string{
			aws.String("a*"),
			aws.String("c*"),
			aws.String("d*"),
			aws.String("h*"),
			aws.String("i*"),
			aws.String("m*"),
			aws.String("r*"),
			aws.String("t*"),
			aws.String("z*"),
		},
	}
}

func activeInstanceFilter() *ec2.Filter {
	return &ec2.Filter{
		Name: aws.String("instance-state-name"),
		Values: []*string{
			aws.String("pending"),
			aws.String("running"),
		},
	}
}

// standardInstancesCPUs returns the number of vCPUs for all standard
// (A, C, D, H, I, M, R, T, Z) EC2 instances
// Note that we are working out the number of vCPUs for each instance
// here because instances can have custom CPU options specified during
// launch. More information can be found at
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html
func standardInstancesCPUs(ec2Service ec2iface.EC2API, spotInstances bool) (int64, error) {
	var totalvCPUs int64
	instanceTypeFilter := standardInstanceTypeFilter()
	instanceStateFilter := activeInstanceFilter()
	filters := []*ec2.Filter{instanceTypeFilter, instanceStateFilter}

	// According to the AWS docs we should be able to filter
	// "scheduled" instances as well, but that does not work so we
	// are using filters only for the spot instances
	if spotInstances {
		spotFilter := &ec2.Filter{
			Name:   aws.String("instance-lifecycle"),
			Values: []*string{aws.String("spot")},
		}
		filters = append(filters, spotFilter)
	}

	params := &ec2.DescribeInstancesInput{Filters: filters}
	err := ec2Service.DescribeInstancesPages(params,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			if page != nil {
				for _, reservation := range page.Reservations {
					for _, instance := range reservation.Instances {
						// InstanceLifecycle is nil for On-Demand instances
						if !spotInstances && instance.InstanceLifecycle != nil {
							continue
						}

						cpuOptions := instance.CpuOptions
						if cpuOptions.CoreCount != nil && cpuOptions.ThreadsPerCore != nil {
							numvCPUs := *cpuOptions.CoreCount * *cpuOptions.ThreadsPerCore
							totalvCPUs += numvCPUs
						}
					}
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return 0, err
	}

	return totalvCPUs, nil
}

// StandardSpotInstanceRequestsUsageCheck implements the UsageCheck interface
// for standard spot instance requests
type StandardSpotInstanceRequestsUsageCheck struct {
	client      ec2iface.EC2API
}

// Usage returns vCPU usage for all standard (A, C, D, H, I, M, R, T,
// Z) spot instance requests and usage or an error
// vCPUs are returned instead of the number of images due to the
// service quota reporting the number of vCPUs
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-limits.html
func (c *StandardSpotInstanceRequestsUsageCheck) Usage() ([]QuotaUsage, error) {
	cpus, err := standardInstancesCPUs(c.client, true)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	usage := []QuotaUsage{
		{
			Name:        spotInstanceRequestsName,
			Description: spotInstanceRequestsDesc,
			Usage:       float64(cpus),
		},
	}
	return usage, nil
}

// RunningOnDemandStandardInstancesUsageCheck implements the UsageCheck interface
// for standard on-demand instances
type RunningOnDemandStandardInstancesUsageCheck struct {
	client      ec2iface.EC2API
}

// Usage returns vCPU usage for all running on-demand standard (A, C,
// D, H, I, M, R, T, Z) instances or an error vCPUs are returned instead
// of the number of images due to the service quota reporting the number
// of vCPUs
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-limits.html
func (c *RunningOnDemandStandardInstancesUsageCheck) Usage() ([]QuotaUsage, error) {
	cpus, err := standardInstancesCPUs(c.client, false)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	usage := []QuotaUsage{
		{
			Name:        onDemandInstanceRequestsName,
			Description: onDemandInstanceRequestsDesc,
			Usage:       float64(cpus),
		},
	}
	return usage, nil
}
