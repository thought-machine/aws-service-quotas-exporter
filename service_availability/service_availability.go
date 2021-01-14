package serviceavailability

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

// Errors returned from this package
var (
	ErrInvalidRegion             = errors.New("invalid region")
	ErrFailedToGetSubnet         = errors.New("failed to get subnet")
	ErrFailedToGetIpAvailability = errors.New("failed to get IP availability")
)

var services = []string{"ec2", "vpc"}

// Can probably skip making this struct and just use the default
// AWS output type "subnet" if we don't want to expand this in
// the future? Although to be honest I don't really see how this
// can be easily expanded as this API is unique?
// If we want more subnet info in the future we can just keep all
// of it from the default output and add/remove filters here later?

// AvailabilityInfo reprents the subnet IP availability and other info?
type AvailabilityInfo struct {
	// Name is the name given to the piece of exported availibility 
	// information (eg. available_IPs_per_subnet)
	Name string
	// Description is the name given to the piece of exported 
	// availibility information (eg. available IPs per subnet)
	Description string
	// 
	SubnetArn *string
	//
	CidrBlock *string
	//
	AvailableIps int
	//
	Limit int
}

