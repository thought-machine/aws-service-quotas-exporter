package servicequotas

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func (m *mockAutoScalingClient) DescribeAutoScalingGroupsPages(input *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error {
	fn(m.DescribeAutoScalingGroupsResponse, true)
	return m.err
}

func TestASGUsageCheckWithError(t *testing.T) {
	mockClient := &mockAutoScalingClient{
		err:                               errors.New("some err"),
		DescribeAutoScalingGroupsResponse: nil,
	}

	check := ASGUsageCheck{mockClient}
	usage, err := check.Usage()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFailedToGetUsage))
	assert.Nil(t, usage)
}

func TestASGUsageCheck(t *testing.T) {
	mockClient := &mockAutoScalingClient{
		err: errors.New("some err"),
		DescribeAutoScalingGroupsResponse: autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg1"),
					Instances:            []*autoscaling.Instance{
						{LifecycleState: aws.String("Terminating")},
						{LifecycleState: aws.String("Terminating:Wait")},
						{LifecycleState: aws.String("Terminating:Proceed")},
						{LifecycleState: aws.String("Terminated")},
						{LifecycleState: aws.String("Detaching")},
						{LifecycleState: aws.String("Detached")},
						{LifecycleState: aws.String("InService")},
						{LifecycleState: aws.String("Pending")},
					},
					MaxSize:              aws.Int64(7),
				},
				{
					AutoScalingGroupName: aws.String("asg2"),
					Instances:            []*autoscaling.Instance{},
					MaxSize:              aws.Int64(3),
				},
				{
					AutoScalingGroupName: aws.String("asg3"),
					Instances:            []*autoscaling.Instance{
						{LifecycleState: aws.String("InService")},
						{LifecycleState: aws.String("InService")},
						{LifecycleState: aws.String("Pending")},
					},
					MaxSize:              aws.Int64(10),
				}
			},
		},
	}

	check := ASGUsageCheck{mockClient}
	usage, err := check.Usage()

	expectedUsage := []QuotaUsage{
		{
			Name:         numInstancesPerASGName,
			ResourceName: "asg1",
			Description:  numInstancesPerASGDescription,
			Usage:        float64(2),
			Quota:        float64(7),
		},
		{
			Name:         numInstancesPerASGName,
			ResourceName: "asg2",
			Description:  numInstancesPerASGDescription,
			Usage:        float64(0),
			Quota:        float64(3),
		},
		{
			Name:         numInstancesPerASGName,
			ResourceName: "asg3",
			Description:  numInstancesPerASGDescription,
			Usage:        float64(3),
			Quota:        float64(10),
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedUsage, usage)
}
