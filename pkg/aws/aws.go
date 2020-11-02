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
func (m *Manager) IsSSHKeyUploaded(desiredSSHKey core.SSHKey) (keyID string, err error) {

	// TODO: API calls are region specific. so if import key, only for that region of AWS

	svc := ec2.New(m.api)
	dkpOutput, err := svc.DescribeKeyPairs(nil)
	if err != nil {
		// errrors.New("Unable to get key pairs, %v", err)
		return "", err
	}
	// func (c *EC2) CreateKeyPair(input *CreateKeyPairInput) (*CreateKeyPairOutput, error)
	foundKeyPair := false
	for _, keyPair := range dkpOutput.KeyPairs {
		fmt.Println(keyPair)
		keyID = *keyPair.KeyPairId
	}

	if foundKeyPair {
		return keyID, nil
	}

	// need to create a new keypair
	var keyPairInfo ec2.ImportKeyPairInput
	keyPairInfo.SetKeyName("athir")
	keyPairInfo.SetPublicKeyMaterial([]byte("dakjfdalkjfadslkj"))
	//keyPairInfo.SetDryRun()
	err = keyPairInfo.Validate()
	if err != nil {
		return keyID, err
	}

	keyPair, err := svc.ImportKeyPair(&keyPairInfo)
	if err != nil {
		return keyID, err
	}
	keyID = *keyPair.KeyPairId
	fmt.Println(keyPair.KeyFingerprint)

	return keyID, nil
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
	// var regions []regionPingTimes
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

		// regions = append(regions, region)
		// regions has unordered list of all ping times
	}

	return closestRegion, nil
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID int) (vmInfo core.VMInfo, err error) {

	return vmInfo, nil
}

// CreateVM will create a new VM
func (m *Manager) CreateVM(name string, image string, size string, region string, sshKey core.SSHKey) (core.VMInfo, error) {
	var vmInfo core.VMInfo

	svc := ec2.New(m.api)

	// us-east-1
	// ubuntu 20.04LTS: ami-0dba2cb6798deb6d8 (amd64) ami-0ea142bd244023692 (arm)
	// t1.micro, t2.micro
	// VPC - create new one?
	// security group
	// shutdown behavior

	// Specify the details of the instance that you want to create.
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
		ImageId:      aws.String("ami-e7527ed7"),
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
	})
	fmt.Println(runResult)
	if err != nil {
		fmt.Println("Could not create instance", err)
		return vmInfo, err
	}

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
