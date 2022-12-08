package servicequotas

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
)

type mockAutoScalingClient struct {
	autoscalingiface.AutoScalingAPI

	err                               error
	DescribeAutoScalingGroupsResponse *autoscaling.DescribeAutoScalingGroupsOutput
}
