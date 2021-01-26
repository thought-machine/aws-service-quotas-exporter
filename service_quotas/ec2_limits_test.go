package servicequotas

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func (m *mockEC2Client) DescribeSecurityGroupsPages(input *ec2.DescribeSecurityGroupsInput, fn func(*ec2.DescribeSecurityGroupsOutput, bool) bool) error {
	fn(m.DescribeSecurityGroupsResponse, true)
	return m.err
}

func (m *mockEC2Client) DescribeNetworkInterfacesPages(input *ec2.DescribeNetworkInterfacesInput, fn func(*ec2.DescribeNetworkInterfacesOutput, bool) bool) error {
	fn(m.DescribeNetworkInterfacesResponse, true)
	return m.err
}

func (m *mockEC2Client) DescribeInstancesPages(input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	m.InstancesFilters = input.Filters
	fn(m.DescribeInstancesResponse, true)
	return m.err
}

func (m *mockEC2Client) DescribeSubnetsPages(input *ec2.DescribeSubnetsInput, fn func(*ec2.DescribeSubnetsOutput, bool) bool) error {
	fn(m.DescribeSubnetsResponse, true)
	return m.err
}

func TestRulesPerSecurityGroupUsageWithError(t *testing.T) {
	mockClient := &mockEC2Client{
		err:                            errors.New("some err"),
		DescribeSecurityGroupsResponse: nil,
	}

	check := RulesPerSecurityGroupUsageCheck{mockClient}
	usage, err := check.Usage()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetUsage))
	assert.Nil(t, usage)
}

func TestRulesPerSecurityGroupUsage(t *testing.T) {
	testCases := []struct {
		name           string
		securityGroups []*ec2.SecurityGroup
		expectedUsage  []QuotaUsage
	}{
		{
			name:           "WithNoSecurityGroups",
			securityGroups: []*ec2.SecurityGroup{},
			expectedUsage:  []QuotaUsage{},
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
			expectedUsage: []QuotaUsage{
				{
					Name:         rulesPerSecGrpName,
					ResourceName: aws.String("somegroupid"),
					Description:  rulesPerSecGrpDesc,
					Usage:        0,
				},
				{
					Name:         rulesPerSecGrpName,
					ResourceName: aws.String("groupwithrules"),
					Description:  rulesPerSecGrpDesc,
					Usage:        2,
				},
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

			check := RulesPerSecurityGroupUsageCheck{mockClient}
			usage, err := check.Usage()

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

	check := SecurityGroupsPerENIUsageCheck{mockClient}
	usage, err := check.Usage()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetUsage))
	assert.Nil(t, usage)
}

func TestSecurityGroupsPerENIUsage(t *testing.T) {
	testCases := []struct {
		name              string
		networkInterfaces []*ec2.NetworkInterface
		expectedUsage     []QuotaUsage
	}{
		{
			name:              "WithNoNetworkInterfaces",
			networkInterfaces: []*ec2.NetworkInterface{},
			expectedUsage:     []QuotaUsage{},
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
			expectedUsage: []QuotaUsage{
				{
					Name:         secGroupsPerENIName,
					ResourceName: aws.String("someeni"),
					Description:  secGroupsPerENIDesc,
					Usage:        2,
				},
			},
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

			check := SecurityGroupsPerENIUsageCheck{mockClient}
			usage, err := check.Usage()

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

	check := SecurityGroupsPerRegionUsageCheck{mockClient}
	usage, err := check.Usage()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetUsage))
	assert.Nil(t, usage)
}

func TestSecurityGroupsPerRegionUsage(t *testing.T) {
	testCases := []struct {
		name           string
		securityGroups []*ec2.SecurityGroup
		expectedUsage  []QuotaUsage
	}{
		{
			name:           "WithNoSecurityGroups",
			securityGroups: []*ec2.SecurityGroup{},
			expectedUsage: []QuotaUsage{
				{
					Name:        securityGroupsPerRegionName,
					Description: securityGroupsPerRegionDesc,
					Usage:       0,
				},
			},
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
			expectedUsage: []QuotaUsage{
				{
					Name:        securityGroupsPerRegionName,
					Description: securityGroupsPerRegionDesc,
					Usage:       2,
				},
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

			check := SecurityGroupsPerRegionUsageCheck{mockClient}
			usage, err := check.Usage()

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUsage, usage)
		})
	}
}

func TestStandardInstancesCPUsWithError(t *testing.T) {
	mockClient := &mockEC2Client{
		err:                       errors.New("some err"),
		DescribeInstancesResponse: nil,
	}

	cpus, err := standardInstancesCPUs(mockClient, true)

	assert.Error(t, err)
	assert.Equal(t, int64(0), cpus)
}

func TestStandardInstancesCPUsFilters(t *testing.T) {
	instanceTypeFilter := standardInstanceTypeFilter()
	instanceStateFilter := activeInstanceFilter()

	testCases := []struct {
		name            string
		spotInstances   bool
		expectedFilters []*ec2.Filter
	}{
		{
			name:          "ForSpotInstances",
			spotInstances: true,
			expectedFilters: []*ec2.Filter{
				instanceTypeFilter,
				instanceStateFilter,
				{
					Name:   aws.String("instance-lifecycle"),
					Values: []*string{aws.String("spot")},
				},
			},
		},
		{
			name:            "ForOnDemandInstances",
			spotInstances:   false,
			expectedFilters: []*ec2.Filter{instanceTypeFilter, instanceStateFilter},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockEC2Client{err: nil, DescribeInstancesResponse: nil}

			cpus, err := standardInstancesCPUs(mockClient, tc.spotInstances)

			assert.NoError(t, err)
			assert.Equal(t, int64(0), cpus)
			assert.Equal(t, mockClient.InstancesFilters, tc.expectedFilters)
		})
	}
}

func TestStandardInstancesCPUs(t *testing.T) {
	mockClient := &mockEC2Client{
		err: nil,
		DescribeInstancesResponse: &ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances: []*ec2.Instance{
						{
							InstanceLifecycle: aws.String("spot"),
							CpuOptions: &ec2.CpuOptions{
								CoreCount:      aws.Int64(4),
								ThreadsPerCore: aws.Int64(2),
							},
						},
					},
				},
				{
					Instances: []*ec2.Instance{
						{
							CpuOptions: &ec2.CpuOptions{
								CoreCount:      aws.Int64(2),
								ThreadsPerCore: aws.Int64(2),
							},
						},
						{
							CpuOptions: &ec2.CpuOptions{
								CoreCount:      aws.Int64(4),
								ThreadsPerCore: aws.Int64(2),
							},
						},
					},
				},
			},
		},
	}

	cpus, err := standardInstancesCPUs(mockClient, false)
	assert.NoError(t, err)
	assert.Equal(t, int64(12), cpus)
}

func TestAvailableIpsPerSubnetUsageWithError(t *testing.T) {
	mockClient := &mockEC2Client{
		err:                     errors.New("some err"),
		DescribeSubnetsResponse: nil,
	}

	check := AvailableIpsPerSubnetUsageCheck{mockClient}
	usage, err := check.Usage()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetAvailability))
	assert.Nil(t, usage)
}

func TestAvailableIpsPerSubnetUsageWithInvalidCidrConversion(t *testing.T) {
	mockClient := &mockEC2Client{
		DescribeSubnetsResponse: &ec2.DescribeSubnetsOutput{
			Subnets: []*ec2.Subnet{
				{
					AvailabilityZone:        aws.String("eu-west-1"),
					AvailableIpAddressCount: aws.Int64(4096),
					CidrBlock:               aws.String("invalid-cidr"),
					SubnetArn:               aws.String("subnet-arn"),
				},
			},
		},
	}
	check := AvailableIpsPerSubnetUsageCheck{mockClient}
	usage, err := check.Usage()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToConvertCidr))
	assert.Nil(t, usage)
}

func TestAvailableIpsPerSubnetUsage(t *testing.T) {
	testCases := []struct {
		name          string
		subnets       []*ec2.Subnet
		expectedUsage []QuotaUsage
	}{
		{
			name:          "WithNoSubnets",
			subnets:       []*ec2.Subnet{},
			expectedUsage: []QuotaUsage{},
		},
		{
			name: "WithSingleSubnet",
			subnets: []*ec2.Subnet{
				{
					AvailabilityZone:        aws.String("eu-west-1"),
					AvailableIpAddressCount: aws.Int64(4096),
					CidrBlock:               aws.String("100.10.10.0/20"),
					SubnetArn:               aws.String("subnet-arn"),
				},
			},
			expectedUsage: []QuotaUsage{
				{
					Name:         availableIPsPerSubnetName,
					ResourceName: aws.String("subnet-arn"),
					Description:  availableIPsPerSubnetDesc + "100.10.10.0/20",
					Usage:        float64(0),
					Quota:        float64(4096),
				},
			},
		},
		{
			name: "WithMultipleSubnets",
			subnets: []*ec2.Subnet{
				{
					AvailabilityZone:        aws.String("eu-west-1"),
					AvailableIpAddressCount: aws.Int64(4096),
					CidrBlock:               aws.String("100.10.10.0/20"),
					SubnetArn:               aws.String("subnet-arn-1"),
				},
				{
					AvailabilityZone:        aws.String("eu-west-1"),
					AvailableIpAddressCount: aws.Int64(0),
					CidrBlock:               aws.String("100.10.10.0/21"),
					SubnetArn:               aws.String("subnet-arn-2"),
				},
				{
					AvailabilityZone:        aws.String("eu-west-1"),
					AvailableIpAddressCount: aws.Int64(1024),
					CidrBlock:               aws.String("100.10.10.0/22"),
					SubnetArn:               aws.String("subnet-arn-3"),
				},
			},
			expectedUsage: []QuotaUsage{
				{
					Name:         availableIPsPerSubnetName,
					ResourceName: aws.String("subnet-arn-1"),
					Description:  availableIPsPerSubnetDesc + "100.10.10.0/20",
					Usage:        float64(0),
					Quota:        float64(4096),
				},
				{
					Name:         availableIPsPerSubnetName,
					ResourceName: aws.String("subnet-arn-2"),
					Description:  availableIPsPerSubnetDesc + "100.10.10.0/21",
					Usage:        float64(1),
					Quota:        float64(2048),
				},
				{
					Name:         availableIPsPerSubnetName,
					ResourceName: aws.String("subnet-arn-3"),
					Description:  availableIPsPerSubnetDesc + "100.10.10.0/22",
					Usage:        float64(0),
					Quota:        float64(1024),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				err: nil,
				DescribeSubnetsResponse: &ec2.DescribeSubnetsOutput{
					Subnets: tc.subnets,
				},
			}

			check := AvailableIpsPerSubnetUsageCheck{mockClient}
			usage, err := check.Usage()

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUsage, usage)
		})
	}
}
