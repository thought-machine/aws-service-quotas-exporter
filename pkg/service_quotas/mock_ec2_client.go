package servicequotas

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type mockEC2Client struct {
	ec2iface.EC2API

	err                               error
	DescribeSecurityGroupsResponse    *ec2.DescribeSecurityGroupsOutput
	DescribeNetworkInterfacesResponse *ec2.DescribeNetworkInterfacesOutput
	InstancesFilters                  []*ec2.Filter
	DescribeInstancesResponse         *ec2.DescribeInstancesOutput
	DescribeSubnetsResponse           *ec2.DescribeSubnetsOutput
}
