package servicequotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

const serviceName = "vpc"

// Errors returned from this package
var (
	ErrFailedToGetUsage = errors.New("failed to get usage")
)

var newEC2Service = ec2New

func ec2New(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
	return ec2.New(c, cfgs...)
}

// All the usage limits checks in this file are reported under the
// `vpc` `ServiceCode` by the AWS Service Quotas

// RulesPerSecurityGroupUsage checks the usage for the "Inbound or
// outbound rules per security group" quota
type RulesPerSecurityGroupUsage struct {
}

// Code is the AWS service quotas code
func (u *RulesPerSecurityGroupUsage) Code() string {
	return "L-0EA8095F"
}

// Name is the name of the service quota
func (u *RulesPerSecurityGroupUsage) Name() string {
	return "Inbound or outbound rules per security group"
}

// ServiceName is the name of the service as reported by the service
// quotas service
func (u *RulesPerSecurityGroupUsage) ServiceName() string {
	return serviceName
}

// Usage returns a map of security group IDs and the sum of their
// inbound and outbound rules or an error
func (u *RulesPerSecurityGroupUsage) Usage(c client.ConfigProvider, cfgs ...*aws.Config) (map[string]float64, error) {
	usage := map[string]float64{}

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
		usage[*securityGroup.GroupId] = float64(inboundRules + outboundRules)
	}

	return usage, nil
}

// SecurityGroupsPerENIUsage checks the usage for the "Security groups
// per network interface" quota
type SecurityGroupsPerENIUsage struct {
}

// Code is the AWS service quotas code
func (u *SecurityGroupsPerENIUsage) Code() string {
	return "L-2AFB9258"
}

// Name is the name of the service quota
func (u *SecurityGroupsPerENIUsage) Name() string {
	return "Security groups per network interface"
}

// ServiceName is the name of the service as reported by the service
// quotas service
func (u *SecurityGroupsPerENIUsage) ServiceName() string {
	return serviceName
}

// Usage returns a map of Elastic Network Interface IDs and the number
// of security groups for each network interface or an error
func (u *SecurityGroupsPerENIUsage) Usage(c client.ConfigProvider, cfgs ...*aws.Config) (map[string]float64, error) {
	usage := map[string]float64{}

	ec2Service := newEC2Service(c, cfgs...)
	params := &ec2.DescribeNetworkInterfacesInput{}
	err := ec2Service.DescribeNetworkInterfacesPages(params,
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			if page != nil {
				for _, eni := range page.NetworkInterfaces {
					numSecurityGroups := len(eni.Groups)
					usage[*eni.NetworkInterfaceId] = float64(numSecurityGroups)
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetUsage, "%w", err)
	}

	return usage, nil
}

// SecurityGroupsPerRegionUsage checks the usage for the "VPC security
// groups per Region" quota
type SecurityGroupsPerRegionUsage struct {
}

// Code is the AWS service quotas code
func (u *SecurityGroupsPerRegionUsage) Code() string {
	return "L-E79EC296"
}

// Name is the name of the service quota
func (u *SecurityGroupsPerRegionUsage) Name() string {
	return "VPC security groups per Region"
}

// ServiceName is the name of the service as reported by the service
// quotas service
func (u *SecurityGroupsPerRegionUsage) ServiceName() string {
	return serviceName
}

// Usage returns a map of the service quota name for that service (see
// the `Name` method) and the number of security groups or an error
func (u *SecurityGroupsPerRegionUsage) Usage(c client.ConfigProvider, cfgs ...*aws.Config) (map[string]float64, error) {
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

	usage := map[string]float64{u.Name(): float64(numGroups)}
	return usage, nil
}
