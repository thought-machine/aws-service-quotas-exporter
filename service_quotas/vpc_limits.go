package service_quotas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

const serviceName = "vpc"

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

func (u *RulesPerSecurityGroupUsage) Usage(awsSession *session.Session, cfgs ...*aws.Config) (map[string]float64, error) {
	return nil, nil
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

func (u *SecurityGroupsPerENIUsage) Usage(awsSession *session.Session, cfgs ...*aws.Config) (map[string]float64, error) {
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

func (u *SecurityGroupsPerRegionUsage) Usage(awsSession *session.Session, cfgs ...*aws.Config) (map[string]float64, error) {
	return nil, nil
}
