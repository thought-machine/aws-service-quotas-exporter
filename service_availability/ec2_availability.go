package serviceavailability

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

const (
	availableIPsPerSubnetName = "available_IPs_per_subnet"
	availableIPsPerSubnetDesc = "available IPs per subnet"
)

type AvailableIpsPerSubnetCheck struct {
	client ec2iface.EC2API
}

// Availability returns the usage for each subnet ID with the usage value
// being the number of available IPv4 addresses in that subnet or
// an error
func (c *AvailableIpsPerSubnetCheck) Availability() ([]AvailabilityInfo, error) {
	availabilityInfos := []AvailabilityInfo{}

	params := &ec2.DescribeSubnetsInput{} // do we need dryrun false flag here?
	err := c.client.DescribeSubnetsPages(params,
		func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			if page != nil {
				for _, subnet := range page.Subnets {
					availabilityInfo := AvailabilityInfo{
						Name:         availableIPsPerSubnetName,
						Description:  availableIPsPerSubnetDesc,
						SubnetArn:    subnet.SubnetArn,
						CidrBlock:    subnet.CidrBlock,
						AvailableIps: float64(*subnet.AvailableIpAddressCount),
					}
					availabilityInfos = append(availabilityInfos, availabilityInfo)
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ErrFailedToGetIpAvailability, "%w", err)
	}

	return availabilityInfos, nil
}
