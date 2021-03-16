package servicequotas

import (
	"testing"

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
