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

# IAM Policy required for the exporter to run

TODO:
Describe security groups
Describe network interfaces
Describe instances
service quotas - list service quotas

```
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

# Adding additional metrics

1. Implement the `QuotasInterface`.
   * Return a slice of `QuotaUsage` with the `ResourceName` if the
     check is for multiple resources, such as the rules per security
     group or a slice with a single element and the `ResourceName` set
     to `nil` for checks like the spot instance requests.
2. Add the check to the `newUsageChecks` and make sure to pass the appropriate AWS client
3. If the service for that usage check is not present in `services` (`service_quotas/service_quotas.go`), add it.

[1]: https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html
[2]: https://prometheus.io/
[3]: https://aws.amazon.com/premiumsupport/plans/
[4]: https://docs.aws.amazon.com/awssupport/latest/APIReference/Welcome.html
