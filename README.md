# AWS Service Quotas Exporter
The aws-service-quotas-exporter exports [AWS service quotas][1] and
usage as [Prometheus][2] metrics. This exporter only uses the service
quotas API and has custom implementation for each usage metric.
That makes it suitable for AWS accounts that do not have [Business or Enterprise
support plan][3], required by the [AWS Support API][4] (AWS
Trusted Advisor). This exporter also provides some metrics that are
not available via the AWS Trusted Advisor, such as "rules per security
group" and "spot instance requests". Other metrics exported through other AWS APIs
can also be integrated with minimal effort, an example of such is "available IPs
per subnet" as seen in //service_quotas/ec2_limits.go.

# Metrics

There are 7 metrics exposed:

1. Rules per security group
```
aws_inbound_rules_per_security_group_limit_total{region="eu-west-1",resource="sg-0000000000000"} 200
aws_inbound_rules_per_security_group_used_total{region="eu-west-1",resource="sg-0000000000000"} 198
aws_outbound_rules_per_security_group_limit_total{region="eu-west-1",resource="sg-00000000000000"} 200
aws_outbound_rules_per_security_group_used_total{region="eu-west-1",resource="sg-00000000000000"} 7
```

2. Security groups per network interface
```
aws_security_groups_per_network_interface_limit_total{region="eu-west-1",resource="eni-00000000000"} 5
aws_security_groups_per_network_interface_used_total{region="eu-west-1",resource="eni-00000000000"} 1
```

3. Security groups per region
```
aws_security_groups_per_region_limit_total{region="eu-west-1",resource="security_groups_per_region"} 2500
aws_security_groups_per_region_used_total{region="eu-west-1",resource="security_groups_per_region"} 108
```

4. Spot instance requests
```
aws_spot_instance_requests_limit_total{region="eu-west-1",resource="spot_instance_requests"} 640
aws_spot_instance_requests_used_total{region="eu-west-1",resource="spot_instance_requests"} 472
```

5. On-demand instance requests
```
aws_ondemand_instance_requests_limit_total{region="eu-west-1",resource="ondemand_instance_requests"} 9088
aws_ondemand_instance_requests_used_total{region="eu-west-1",resource="ondemand_instance_requests"} 440
```

6. Available IPs per subnet
```
aws_available_ips_per_subnet_limit_total{region="eu-west-1",resource="subnet-do93c3jpg5oe4txjn"} 8192
aws_available_ips_per_subnet_used_total{region="eu-west-1",resource="subnet-do93c3jpg5oe4txjn"} 7959
```

7. VMs per AutoScalingGroup - useful to get alerts if the max number of instances for an ASG has been reached
```
aws_instances_per_asg_limit_total{region="eu-west-1",resource="asg"} 5
aws_instances_per_asg_used_total{region="eu-west-1",resource="asg"} 10
```

# IAM Permissions

The AWS Service Quotas requires permissions for the following actions
to be able to run:

 * `ec2:DescribeSecurityGroups`
 * `ec2:DescribeNetworkInterfaces`
 * `ec2:DescribeInstances`
 * `ec2:DescribeSubnets`
 * `servicequotas:ListServiceQuotas`
 * `autoscaling:DescribeAutoScalingGroups`

Example IAM policy
```
{
   "Version": "2012-10-17",
   "Statement": [{
      "Effect": "Allow",
      "Action": [
          "ec2:DescribeSecurityGroups",
          "ec2:DescribeNetworkInterfaces",
          "ec2:DescribeInstances",
          "ec2:DescribeSubnets",
          "servicequotas:ListServiceQuotas",
          "autoscaling:DescribeAutoScalingGroups"
      ],
      "Resource": "*"
   }]
}
```

# Options

`plz run //cmd:aws-service-quotas-exporter -- [OPTIONS]`
| Short Flag | Long Flag | Env var     | Description                    |
|------------|-----------|-------------|--------------------------------|
| -p         | --port    | N/A         | Port on which to serve metrics |
| -r         | --region  | AWS_REGION  | AWS region                     |
| -f         | --profile | AWS_PROFILE | Named AWS profile              |

# Building the exporter and running the exporter

## Building the binary with please
`plz build //cmd:aws-service-quotas-exporter`

`plz run //cmd:aws-service-quotas-exporter -- -p 9090 -r eu-west-1 --profile myprofile`

## Docker image
`docker build -f build/Dockerfile-builder . --rm=false`

Docker images are also available at thoughtmachine/aws-service-quotas-exporter:<version> See https://hub.docker.com/r/thoughtmachine/aws-service-quotas-exporter

# Extending the exporter with additional metrics

### Implement the `QuotasInterface`.

Example
`service_quotas/<service_name>_limits.go`
```
const (
    myQuotaName        = "prometheus_valid_metric_name"  // Only [a-zA-Z0-9:_]
    myQuotaDescription = "my description"
)

type MyUsageCheck struct {
    client awsserviceiface.SERVICEAPI  // eg ec2iface.EC2API
}

func (c *MyUsageCheck) Usage() ([]QuotaUsage, error) {
    // ...client.GetRequiredInformation

    // In case we are retrieving usage for multiple resources:
    for _, resource := range {
        usage := QuotaUsage{
            Name:         myQuotaName,
            ResourceName: resource.Identifier,
            Description:  myQuotaDescription,
            Usage:        myUsage,
        }
        usages = append(usages, usage)
    }

    // For a single resource
    usages := []QuotaUsage{
        {
            Name:        myQuotaName,
            Description: myQuotaDescription,
            Usage:       myUsage,
        },
    }

    return usages, err
}
```

### Add the check to the `newUsageChecks` and make sure to pass the appropriate AWS client

If the check uses the Service Quotas API, then it needs to be added as part of
`serviceQuotasUsageChecks` with its service quota code (examples given in the
[using AWS CLI to manage service quota requests page][5]). Otherwise, the check can
just be added to `otherUsageChecks`.

`service_quotas/service_quotas.go`
```
func newUsageChecks(c client.ConfigProvider, cfgs ...*aws.Config) map[string]UsageCheck {
    myClient := someawsclient.New(c, cfgs)

    serviceQuotasUsageChecks := map[string]UsageCheck{
        //... other usage checks
        "L-SERVICE_QUOTAS_CODE": &MyUsageCheck{myClient},
    }

    otherUsageChecks := []UsageCheck{
        &MyOtherUsageCheck{ec2Client},
    }

    return serviceQuotasUsageChecks, otherUsageChecks
}
```

### Update this README with the required actions :) (See the IAM Permissions section)


[1]: https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html
[2]: https://prometheus.io/
[3]: https://aws.amazon.com/premiumsupport/plans/
[4]: https://docs.aws.amazon.com/awssupport/latest/APIReference/Welcome.html
[5]: https://aws.amazon.com/premiumsupport/knowledge-center/troubleshoot-service-quotas-cli-commands/
