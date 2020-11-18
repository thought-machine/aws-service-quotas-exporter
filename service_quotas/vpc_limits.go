package service_quotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

const serviceName = "vpc"

var (
	ErrFailedToGetUsage = errors.New("failed to get usage")
)

var newEC2Service = ec2New

func ec2New(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
	return ec2.New(c, cfgs...)
}

// RulesPerSecurityGroupUsage checks the usage for the "Inbound or
// outbound rules per security group" quota
type RulesPerSecurityGroupUsage struct {
}

func (u *RulesPerSecurityGroupUsage) Code() string {
	return "L-0EA8095F"
}

func (u *RulesPerSecurityGroupUsage) Name() string {
	return "Inbound or outbound rules per security group"
}

func (u *RulesPerSecurityGroupUsage) ServiceName() string {
	return serviceName
}

func (u *RulesPerSecurityGroupUsage) Usage(c client.ConfigProvider, cfgs ...*aws.Config) (map[string]float64, error) {
	usage := map[string]float64{}

	ec2Service := newEC2Service(c, cfgs...)

	securityGroups := []*ec2.SecurityGroup{}
	params := &ec2.DescribeSecurityGroupsInput{}
	err := ec2Service.DescribeSecurityGroupsPages(params, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		if page != nil {
			for _, group := range page.SecurityGroups {
				securityGroups = append(securityGroups, group)
			}
		}
		return !lastPage
	})
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

func (u *SecurityGroupsPerENIUsage) Code() string {
	return "L-2AFB9258"
}

func (u *SecurityGroupsPerENIUsage) Name() string {
	return "Security groups per network interface"
}

func (u *SecurityGroupsPerENIUsage) ServiceName() string {
	return serviceName
}

func (u *SecurityGroupsPerENIUsage) Usage(c client.ConfigProvider, cfgs ...*aws.Config) (map[string]float64, error) {
	return nil, nil
}

// SecurityGroupsPerRegionUsage checks the usage for the "VPC security
// groups per Region" quota
type SecurityGroupsPerRegionUsage struct {
}

func (u *SecurityGroupsPerRegionUsage) Code() string {
	return "L-E79EC296"
}

func (u *SecurityGroupsPerRegionUsage) Name() string {
	return "VPC security groups per Region"
}

func (u *SecurityGroupsPerRegionUsage) ServiceName() string {
	return serviceName
}

func (u *SecurityGroupsPerRegionUsage) Usage(c client.ConfigProvider, cfgs ...*aws.Config) (map[string]float64, error) {
	return nil, nil
}
