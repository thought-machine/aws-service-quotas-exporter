package servicequotas

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// type mockEC2Client struct {
// 	ec2iface.EC2API

// 	err                     error
// 	DescribeSubnetsResponse *ec2.DescribeSubnetsOutput
// }

func (m *mockEC2Client) DescribeSubnetsPages(input *ec2.DescribeSubnetsInput, fn func(*ec2.DescribeSubnetsOutput, bool) bool) error {
	fn(m.DescribeSubnetsResponse, true)
	return m.err
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
					ResourceName: aws.String("subnet-arn|100.10.10.0/20"),
					Description:  availableIPsPerSubnetDesc,
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
					ResourceName: aws.String("subnet-arn-1|100.10.10.0/20"),
					Description:  availableIPsPerSubnetDesc,
					Usage:        float64(0),
					Quota:        float64(4096),
				},
				{
					Name:         availableIPsPerSubnetName,
					ResourceName: aws.String("subnet-arn-2|100.10.10.0/21"),
					Description:  availableIPsPerSubnetDesc,
					Usage:        float64(1),
					Quota:        float64(2048),
				},
				{
					Name:         availableIPsPerSubnetName,
					ResourceName: aws.String("subnet-arn-3|100.10.10.0/22"),
					Description:  availableIPsPerSubnetDesc,
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
