package aws

import (
	"fmt"
	"math"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/eezhee/eezhee/pkg/core"
)

// Manager controls access to AWS
type Manager struct {
	APIToken string
	api      *session.Session
}

// NewManager creates a manage object & inits it
func NewManager(providerAPIToken string) (m *Manager) {

	// if user as aws-cli installed, sdk can find ~/.aws/credential file and load keys
	// otherwise, we can create it when user enters details

	if len(providerAPIToken) == 0 {
		fmt.Println("no aws api token set")
		return nil
	}

	manager := new(Manager)
	manager.APIToken = providerAPIToken

	manager.api = session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"),
	}))

	// manager.api = session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))

	return manager
}

// IsSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) IsSSHKeyUploaded(fingerprint string) (bool, error) {

	// TODO: can you upload a keypair? seems like it just generates a new one

	svc := ec2.New(m.api)
	keyPairInfo, err := svc.DescribeKeyPairs(nil)
	if err != nil {
		// errrors.New("Unable to get key pairs, %v", err)
		return false, err
	}
	// func (c *EC2) CreateKeyPair(input *CreateKeyPairInput) (*CreateKeyPairOutput, error)
	foundKeyPair := false
	for _, keyPair := range keyPairInfo.KeyPairs {
		fmt.Println(keyPair)
	}

	if foundKeyPair {
		return true, nil
	}

	// need to create a new keypair

	return true, nil
}

type regionPingTimes struct {
	name      string // region name
	ipAddress string // ip address in region that we can use for ping tests
	result    int64  // ping time for given ip address
}

// SelectClosestRegion will check all DO regions to find the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {

	// get a list of regions
	svc := ec2.New(m.api)
	awsRegions, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return "", err
	}

	// test each region
	var lowestPingTime = math.MaxInt64
	var regions []regionPingTimes
	for _, awsRegion := range awsRegions.Regions {

		// OptInStatus // opt-in-not-required, opted-in and not-opted-in
		// Endpoint    // ie ec2.eu-north-1.amazonaws.com

		region := regionPingTimes{name: *awsRegion.RegionName, ipAddress: *awsRegion.Endpoint, result: 0}

		pingTime, err := core.GetPingTime(region.ipAddress)
		if err != nil {
			return "", err
		}
		region.result = pingTime

		// is this datacenter closer than others we've seen so far
		if int(pingTime) < lowestPingTime {
			closestRegion = region.name
			lowestPingTime = int(pingTime)
		}

		regions = append(regions, region)
	}

	return closestRegion, nil
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID int) (vmInfo core.VMInfo, err error) {
	return vmInfo, nil
}

// CreateVM will create a new VM
func (m *Manager) CreateVM(name string, image string, size string, region string, sshFingerprint string) (core.VMInfo, error) {
	var vmInfo core.VMInfo
	return vmInfo, nil
}

// ListVMs will return a list of all VMs created by eezhee
func (m *Manager) ListVMs() (vmInfo []core.VMInfo, err error) {
	return vmInfo, nil
}

// DeleteVM will delete a given VM
func (m *Manager) DeleteVM(ID int) error {
	return nil
}
