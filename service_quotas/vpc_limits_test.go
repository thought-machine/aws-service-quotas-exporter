package servicequotas

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockEC2Client struct {
	ec2iface.EC2API

	err                               error
	DescribeSecurityGroupsResponse    *ec2.DescribeSecurityGroupsOutput
	DescribeNetworkInterfacesResponse *ec2.DescribeNetworkInterfacesOutput
}

func (m *mockEC2Client) DescribeSecurityGroupsPages(input *ec2.DescribeSecurityGroupsInput, fn func(*ec2.DescribeSecurityGroupsOutput, bool) bool) error {
	fn(m.DescribeSecurityGroupsResponse, true)
	return m.err
}

func (m *mockEC2Client) DescribeNetworkInterfacesPages(input *ec2.DescribeNetworkInterfacesInput, fn func(*ec2.DescribeNetworkInterfacesOutput, bool) bool) error {
	fn(m.DescribeNetworkInterfacesResponse, true)
	return m.err
}

func TestRulesPerSecurityGroupUsageWithError(t *testing.T) {
	mockClient := &mockEC2Client{
		err:                            errors.New("some err"),
		DescribeSecurityGroupsResponse: nil,
	}

	origNewEC2Service := newEC2Service
	defer func() { newEC2Service = origNewEC2Service }()
	newEC2Service = func(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
		return mockClient
	}

	usageCheck := &RulesPerSecurityGroupUsage{}
	usage, err := usageCheck.Usage(nil)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetUsage))
	assert.Nil(t, usage)
}

func TestRulesPerSecurityGroupUsage(t *testing.T) {
	testCases := []struct {
		name           string
		securityGroups []*ec2.SecurityGroup
		expectedUsage  map[string]float64
	}{
		{
			name:           "WithNoSecurityGroups",
			securityGroups: []*ec2.SecurityGroup{},
			expectedUsage:  map[string]float64{},
		},
		{
			name: "WithSecurityGroups",
			securityGroups: []*ec2.SecurityGroup{
				{
					GroupId:             aws.String("somegroupid"),
					IpPermissions:       []*ec2.IpPermission{},
					IpPermissionsEgress: []*ec2.IpPermission{},
				},
				{
					GroupId: aws.String("groupwithrules"),
					IpPermissions: []*ec2.IpPermission{
						{
							FromPort: aws.Int64(0),
							ToPort:   aws.Int64(0),
						},
					},
					IpPermissionsEgress: []*ec2.IpPermission{
						{
							FromPort: aws.Int64(0),
							ToPort:   aws.Int64(0),
						},
					},
				},
			},
			expectedUsage: map[string]float64{
				"somegroupid":    float64(0),
				"groupwithrules": float64(2),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				err: nil,
				DescribeSecurityGroupsResponse: &ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: tc.securityGroups,
				},
			}

			origNewEC2Service := newEC2Service
			defer func() { newEC2Service = origNewEC2Service }()
			newEC2Service = func(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
				return mockClient
			}

			usageCheck := &RulesPerSecurityGroupUsage{}
			usage, err := usageCheck.Usage(nil)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUsage, usage)
		})
	}
}

func TestSecurityGroupsPerENIUsageWithError(t *testing.T) {
	mockClient := &mockEC2Client{
		err:                               errors.New("some err"),
		DescribeNetworkInterfacesResponse: nil,
	}

	origNewEC2Service := newEC2Service
	defer func() { newEC2Service = origNewEC2Service }()
	newEC2Service = func(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
		return mockClient
	}

	usageCheck := &SecurityGroupsPerENIUsage{}
	usage, err := usageCheck.Usage(nil)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetUsage))
	assert.Nil(t, usage)
}

func TestSecurityGroupsPerENIUsage(t *testing.T) {
	testCases := []struct {
		name              string
		networkInterfaces []*ec2.NetworkInterface
		expectedUsage     map[string]float64
	}{
		{
			name:              "WithNoNetworkInterfaces",
			networkInterfaces: []*ec2.NetworkInterface{},
			expectedUsage:     map[string]float64{},
		},
		{
			name: "WithNetworkInterfaces",
			networkInterfaces: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("someeni"),
					Groups: []*ec2.GroupIdentifier{
						{
							GroupId:   aws.String("someid"),
							GroupName: aws.String("somename"),
						},
						{
							GroupId:   aws.String("someotherid"),
							GroupName: aws.String("someothername"),
						},
					},
				},
			},
			expectedUsage: map[string]float64{"someeni": float64(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				err: nil,
				DescribeNetworkInterfacesResponse: &ec2.DescribeNetworkInterfacesOutput{
					NetworkInterfaces: tc.networkInterfaces,
				},
			}

			origNewEC2Service := newEC2Service
			defer func() { newEC2Service = origNewEC2Service }()
			newEC2Service = func(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
				return mockClient
			}

			usageCheck := &SecurityGroupsPerENIUsage{}
			usage, err := usageCheck.Usage(nil)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUsage, usage)
		})
	}
}

func TestSecurityGroupsPerRegionUsageWithError(t *testing.T) {
	mockClient := &mockEC2Client{
		err:                            errors.New("some err"),
		DescribeSecurityGroupsResponse: nil,
	}

	origNewEC2Service := newEC2Service
	defer func() { newEC2Service = origNewEC2Service }()
	newEC2Service = func(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
		return mockClient
	}

	usageCheck := &SecurityGroupsPerRegionUsage{}
	usage, err := usageCheck.Usage(nil)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetUsage))
	assert.Nil(t, usage)
}

func TestSecurityGroupsPerRegionUsage(t *testing.T) {
	testCases := []struct {
		name           string
		securityGroups []*ec2.SecurityGroup
		expectedUsage  map[string]float64
	}{
		{
			name:           "WithNoSecurityGroups",
			securityGroups: []*ec2.SecurityGroup{},
			expectedUsage:  map[string]float64{"VPC security groups per Region": 0},
		},
		{
			name: "WithSecurityGroups",
			securityGroups: []*ec2.SecurityGroup{
				{
					GroupId: aws.String("somegroupid"),
				},
				{
					GroupId: aws.String("anothergroupid"),
				},
			},
			expectedUsage: map[string]float64{"VPC security groups per Region": 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				err: nil,
				DescribeSecurityGroupsResponse: &ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: tc.securityGroups,
				},
			}

			origNewEC2Service := newEC2Service
			defer func() { newEC2Service = origNewEC2Service }()
			newEC2Service = func(c client.ConfigProvider, cfgs ...*aws.Config) ec2iface.EC2API {
				return mockClient
			}

			usageCheck := &SecurityGroupsPerRegionUsage{}
			usage, err := usageCheck.Usage(nil)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUsage, usage)
		})
	}
}
