package servicequotas

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/pkg/errors"
)

const (
	numInstancesPerASGName        = "instances_per_asg"
	numInstancesPerASGDescription = "instances per ASG"
)

// ASGUsageCheck implements the UsageCheckInterface for VMs per
// autoscaling group
type ASGUsageCheck struct {
	client autoscalingiface.AutoScalingAPI
}

// Usage returns usage per auto scaling group - the maximum number of
// instances per ASG and the current number of "running" instances per
// ASG.
func (c *ASGUsageCheck) Usage() ([]QuotaUsage, error) {
	quotaUsages := []QuotaUsage{}

	params := &autoscaling.DescribeAutoScalingGroupsInput{}
	err := c.client.DescribeAutoScalingGroupsPages(params,
		func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			if page != nil {
				for _, asg := range page.AutoScalingGroups {
					numRunningInstances := 0
					for _, instance := range asg.Instances {
						if isRunning(instance) {
							numRunningInstances += 1
						}
					}

					quotaUsage := QuotaUsage{
						Name:         numInstancesPerASGName,
						ResourceName: asg.AutoScalingGroupName,
						Description:  numInstancesPerASGDescription,
						Usage:        float64(numRunningInstances),
						Quota:        float64(*asg.MaxSize),
						Tags:         autoscalingTagsToQuotaUsageTags(asg.Tags),
					}
					quotaUsages = append(quotaUsages, quotaUsage)
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

func isRunning(instance *autoscaling.Instance) bool {
	notRunningStates := map[string]bool{
		"Terminating":         true,
		"Terminating:Wait":    true,
		"Terminating:Proceed": true,
		"Terminated":          true,
		"Detaching":           true,
		"Detached":            true,
	}

	_, isNotRunning := notRunningStates[*instance.LifecycleState]
	return !isNotRunning
}

func autoscalingTagsToQuotaUsageTags(tags []*autoscaling.TagDescription) map[string]string {
	length := len(tags)
	if length == 0 {
		return nil
	}

	out := make(map[string]string, length)
	for _, tag := range tags {
		out[ToPrometheusNamingFormat(*tag.Key)] = *tag.Value
	}

	return out
}
