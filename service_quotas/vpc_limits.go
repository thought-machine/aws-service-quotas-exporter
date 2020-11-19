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
	rulesPerSecGrpDesc          = "Inbound or outbound rules per security group"
	secGroupsPerENIDesc         = "Security groups per network interface"
	securityGroupsPerRegionDesc = "VPC security groups per Region"
)

func ec2New(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
	return ec2.New(c, cfgs...)
}

// All the usage limits checks in this file are reported under the
// `vpc` `ServiceCode` by the AWS Service Quotas

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
// for each ENI
func SecurityGroupsPerENIUsage(c client.ConfigProvider, cfgs ...*aws.Config) ([]QuotaUsage, error) {
	quotaUsages := []QuotaUsage{}

	ec2Service := newEC2Service(c, cfgs...)
	params := &ec2.DescribeNetworkInterfacesInput{}
	err := ec2Service.DescribeNetworkInterfacesPages(params,
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			if page != nil {
				for _, eni := range page.NetworkInterfaces {
					usage := QuotaUsage{
						Name: *eni.NetworkInterfaceId,
						Description: secGroupsPerENIDesc,
						Usage: float64(len(eni.Groups)),
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
// with `cfgs`
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
			Name: securityGroupsPerRegionDesc,
			Description: securityGroupsPerRegionDesc,
			Usage: float64(numGroups),
		},
	}
	return usage, nil
}
