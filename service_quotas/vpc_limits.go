package servicequotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

var newEC2Service = ec2New

const (
	rulesPerSecGrpDesc           = "Rules per security group"
	secGroupsPerENIDesc          = "Security groups per network interface"
	securityGroupsPerRegionDesc  = "Security groups per region"
	spotInstanceRequestsDesc     = "Spot instance requests"
	onDemandInstanceRequestsDesc = "On-demand instance requests"
)

func ec2New(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
	return ec2.New(c, cfgs...)
}

// RulesPerSecurityGroupUsage returns the usage for each security
// group ID with the usage value being the sum of their inbound and
// outbound rules or an error
func RulesPerSecurityGroupUsage(c client.ConfigProvider, cfgs ...*aws.Config) ([]QuotaUsage, error) {
	quotaUsages := []QuotaUsage{}

	ec2Service := newEC2Service(c, cfgs...)

	securityGroups := []*ec2.SecurityGroup{}
	params := &ec2.DescribeSecurityGroupsInput{}
	err := ec2Service.DescribeSecurityGroupsPages(params,
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
			Name:        *securityGroup.GroupId,
			Description: rulesPerSecGrpDesc,
			Usage:       float64(inboundRules + outboundRules),
		}
		quotaUsages = append(quotaUsages, quotaUsage)
	}

	return quotaUsages, nil
}

// SecurityGroupsPerENIUsage returns usage for each Elastic Network
// Interface ID with the usage value being the number of security groups
// for each ENI or an error
func SecurityGroupsPerENIUsage(c client.ConfigProvider, cfgs ...*aws.Config) ([]QuotaUsage, error) {
	quotaUsages := []QuotaUsage{}

	ec2Service := newEC2Service(c, cfgs...)
	params := &ec2.DescribeNetworkInterfacesInput{}
	err := ec2Service.DescribeNetworkInterfacesPages(params,
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			if page != nil {
				for _, eni := range page.NetworkInterfaces {
					usage := QuotaUsage{
						Name:        *eni.NetworkInterfaceId,
						Description: secGroupsPerENIDesc,
						Usage:       float64(len(eni.Groups)),
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

// SecurityGroupsPerRegionUsage returns usage for security groups per
// region as the number of all security groups for the region specified
// with `cfgs` or an error
func SecurityGroupsPerRegionUsage(c client.ConfigProvider, cfgs ...*aws.Config) ([]QuotaUsage, error) {
	numGroups := 0

	ec2Service := newEC2Service(c, cfgs...)

	params := &ec2.DescribeSecurityGroupsInput{}
	err := ec2Service.DescribeSecurityGroupsPages(params,
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
			Name:        securityGroupsPerRegionDesc,
			Description: securityGroupsPerRegionDesc,
			Usage:       float64(numGroups),
		},
	}
	return usage, nil
}

// standardInstancesCPUs returns the number of vCPUs for all standard
// (A, C, D, H, I, M, R, T, Z) EC2 instances
// Note that we are working out the number of vCPUs for each instance
// here because instances can have custom CPU options specified during
// launch. More information can be found at
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html
func standardInstancesCPUs(ec2Service ec2iface.EC2API, spotInstances bool) (int64, error) {
	var totalvCPUs int64

	filters := []*ec2.Filter{
		{
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
		},
		{
			Name: aws.String("instance-state-name"),
			Values: []*string{
				aws.String("pending"),
				aws.String("running"),
			},
		},
	}

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
			for _, reservation := range page.Reservations {
				for _, instance := range reservation.Instances {
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
			return !lastPage
		},
	)
	if err != nil {
		return 0, err
	}

	return totalvCPUs, nil
}

// StandardSpotInstanceRequestsUsage returns vCPU usage for all
// standard (A, C, D, H, I, M, R, T, Z) spot instance requests and usage
// or an error
// vCPUs are returned instead of the number of images due to the
// service quota reporting the number of vCPUs
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-limits.html
func StandardSpotInstanceRequestsUsage(c client.ConfigProvider, cfgs ...*aws.Config) ([]QuotaUsage, error) {
	ec2Service := newEC2Service(c, cfgs...)
	cpus, err := standardInstancesCPUs(ec2Service, true)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	usage := []QuotaUsage{
		{
			Name:        spotInstanceRequestsDesc,
			Description: spotInstanceRequestsDesc,
			Usage:       float64(cpus),
		},
	}
	return usage, nil
}

// RunningOnDemandStandardInstancesUsage returns vCPU usage for all running
// on-demand standard (A, C, D, H, I, M, R, T, Z) instances or an error
// vCPUs are returned instead of the number of images due to the
// service quota reporting the number of vCPUs
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-limits.html
func RunningOnDemandStandardInstancesUsage(c client.ConfigProvider, cfgs ...*aws.Config) ([]QuotaUsage, error) {
	ec2Service := newEC2Service(c, cfgs...)
	cpus, err := standardInstancesCPUs(ec2Service, false)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	usage := []QuotaUsage{
		{
			Name:        onDemandInstanceRequestsDesc,
			Description: onDemandInstanceRequestsDesc,
			Usage:       float64(cpus),
		},
	}
	return usage, nil
}
