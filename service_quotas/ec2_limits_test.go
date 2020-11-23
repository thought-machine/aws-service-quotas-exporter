package servicequotas

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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
	InstancesFilters                  []*ec2.Filter
	DescribeInstancesResponse         *ec2.DescribeInstancesOutput
}

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
					Name:        "somegroupid",
					Description: rulesPerSecGrpDesc,
					Usage:       0,
				},
				{
					Name:        "groupwithrules",
					Description: rulesPerSecGrpDesc,
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
					Name:        "someeni",
					Description: secGroupsPerENIDesc,
					Usage:       2,
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
					Name:        securityGroupsPerRegionDesc,
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
					Name:        securityGroupsPerRegionDesc,
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
