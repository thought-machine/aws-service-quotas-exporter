# AWS Service Quotas Exporter
The aws-service-quotas-exporter exports [AWS service quotas][1] and
usage as [Prometheus][2] metrics. This exporter only uses the service
quotas API and has custom implementation for each usage metric.
That makes it suitable for AWS accounts that do not have [Business or Enterprise
support plan][3], required by the [AWS Support API][4] (AWS
Trusted Advisor). This exporter also provides some metrics that are
not available via the AWS Trusted Advisor, such as "rules per security
group" and "spot instance requests".

# Metrics

There are 5 metrics exposed:

1. Rules per security group
```
aws_rules_per_security_group_limit_total{region="eu-west-1",resource="sg-000000000000"} 60
aws_rules_per_security_group_used_total{region="eu-west-1",resource="sg-000000000000"} 3
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

# IAM Permissions

The AWS Service Quotas requires permissions for the following actions
to be able to run:

 * `ec2:DescribeInstances`
 * `cloudwatch:DescribeAlarms`
 * `cloudwatch:GetMetricStatistics`

Example IAM policy
```
{
   "Version": "2012-10-17",
   "Statement": [{
      "Effect": "Allow",
      "Action": [
         "ec2:DescribeInstances",
         "cloudwatch:DescribeAlarms",
         "cloudwatch:GetMetricStatistics"
      ],
      "Resource": "*"
   }
   ]
}
```

# Options

`plz run //:aws-service-quotas-exporter -- [OPTIONS]`
| Short Flag | Long Flag | Env var     | Description                    |
|------------|-----------|-------------|--------------------------------|
| -p         | --port    | N/A         | Port on which to serve metrics |
| -r         | --region  | AWS_REGION  | AWS region                     |
| -f         | --profile | AWS_PROFILE | Named AWS profile              |

# Building the exporter and running the exporter

## Building the binary with please
`plz build //:aws-service-quotas-exporter`

`plz run //:aws-service-quotas-exporter -- -p 9090 -r eu-west-1 --profile myprofile`

## Docker image
`docker build -f Dockerfile-builder . --rm=false`

Docker images are also available at thoughtmachine/aws-service-quotas-exporter:<version> See https://hub.docker.com/r/thoughtmachine/aws-service-quotas-exporter

# Extending the exporter with additional metrics

1. Implement the `QuotasInterface`.

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

2. Add the check to the `newUsageChecks` and make sure to pass the appropriate AWS client

Example:
`service_quotas/service_quotas.go`
```
func newUsageChecks(c client.ConfigProvider, cfgs ...*aws.Config) map[string]UsageCheck {
    myClient := someawsclient.New(c, cfgs)
    return map[string]UsageCheck{
        //... other usage checks
        "L-SERVICE_QUOTAS_CODE": &MyUsageCheck{myClient},
    }
}
```

3. Update this README with the required actions :) (See the IAM Permissions section)


[1]: https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html
[2]: https://prometheus.io/
[3]: https://aws.amazon.com/premiumsupport/plans/
[4]: https://docs.aws.amazon.com/awssupport/latest/APIReference/Welcome.html
