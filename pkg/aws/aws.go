package aws

import (
	"errors"
	"math"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/eezhee/eezhee/pkg/core"
	log "github.com/sirupsen/logrus"
)

// Manager controls access to AWS
type Manager struct {
	APIToken string
	api      *session.Session
}

// NewManager creates a manage object & inits it
func NewManager(providerAPIToken string) (core.VMManager, error) {

	// if user as aws-cli installed, sdk can find ~/.aws/credential file and load keys
	// otherwise, we can create it when user enters details
	manager := new(Manager)

	// make sure we have an api token
	if len(providerAPIToken) == 0 {
		// check places provider CLI tools store token
		providerAPIToken := manager.FindAuthToken()
		if len(providerAPIToken) == 0 {
			// log.Error("no aws api token set")
			return manager, errors.New("no aws api token set")
		}
		// ok we found a token
	}
	manager.APIToken = providerAPIToken

	manager.api = session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"),
	}))

	// manager.api = session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))

	return manager, nil
}

// GetAuthToken will check common place for aws api key
func (m *Manager) FindAuthToken() string {
	return ""
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
		log.Debug(keyPair)
		keyID = *keyPair.KeyPairId
	}

	if foundKeyPair {
		return keyID, nil
	}

	// is not on AWS yet
	return "", err
}

// UploadSSHKey will upload a given ssh key to AWS
func (m *Manager) UploadSSHKey(keyName string, sshKey core.SSHKey) (keyID string, err error) {

	// need to create a new keypair
	var keyPairInfo ec2.ImportKeyPairInput
	keyPairInfo.SetKeyName(keyName)
	keyPairInfo.SetPublicKeyMaterial([]byte(sshKey.GetPublicKey()))
	//keyPairInfo.SetDryRun()
	err = keyPairInfo.Validate()
	if err != nil {
		return keyID, err
	}

	svc := ec2.New(m.api)
	keyPair, err := svc.ImportKeyPair(&keyPairInfo)
	if err != nil {
		return keyID, err
	}
	keyID = *keyPair.KeyPairId
	log.Debug(keyPair.KeyFingerprint)

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
	var lowestPingTime = math.MaxInt32
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
func (m *Manager) GetVMInfo(vmID string) (vmInfo core.VMInfo, err error) {

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
	log.Debug(runResult)
	if err != nil {
		log.Error("Could not create instance", err)
		return vmInfo, err
	}

	return vmInfo, nil
}

// ListVMs will return a list of all VMs created by eezhee
func (m *Manager) ListVMs() (vmInfo []core.VMInfo, err error) {

	return vmInfo, nil
}

// DeleteVM will delete a given VM
func (m *Manager) DeleteVM(ID string) error {

	return nil
}
