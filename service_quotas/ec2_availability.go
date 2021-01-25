package servicequotas

import (
	maths "math"
	"strconv"

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

// Usage returns the usage for each subnet ID with the usage value
// being the number of available IPv4 addresses in that subnet or
// an error
func (c *AvailableIpsPerSubnetCheck) Usage() ([]QuotaUsage, error) {
	availabilityInfos := []QuotaUsage{}

	params := &ec2.DescribeSubnetsInput{}
	err := c.client.DescribeSubnetsPages(params,
		func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			if page != nil {
				for _, subnet := range page.Subnets {
					cidrBlock := *subnet.CidrBlock
					blockedBits, err := strconv.Atoi(cidrBlock[len(cidrBlock)-2:])
					if err != nil {
						continue
					}
					maxNumOfIPs := maths.Pow(2, 32-float64(blockedBits))
					usage := (maxNumOfIPs - float64(*subnet.AvailableIpAddressCount)) / maxNumOfIPs
					resourceName := *subnet.SubnetArn + "|" + *subnet.CidrBlock
					availabilityInfo := QuotaUsage{
						Name:         availableIPsPerSubnetName,
						ResourceName: &resourceName,
						Description:  availableIPsPerSubnetDesc,
						Usage:        float64(usage),
						Quota:        float64(maxNumOfIPs),
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
