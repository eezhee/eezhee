package vultr

import (
	"context"
	"fmt"

	"github.com/eezhee/eezhee/pkg/core"
	"github.com/vultr/govultr"
)

// Manager controls access to AWS
type Manager struct {
	APIToken string
	api      *govultr.Client
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

	manager.api = govultr.NewClient(nil, manager.APIToken)

	return manager
}

// IsSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) IsSSHKeyUploaded(fingerprint string) (bool, error) {

	keys, err := m.api.SSHKey.List(context.Background())
	if err != nil {
		return false, err
	}

	for _, key := range keys {
		fmt.Println(key.Name)
		// check fingerprint and see if we have a match
	}

	// we need to add the key
	newKey, err := m.api.SSHKey.Create(context.Background(), "athir-eezhee", "ssh-rsa dakdfje209u23r")
	if err != nil {
		return false, err
	}
	fmt.Println(newKey.Name)

	return true, nil
}

// SelectClosestRegion will check all DO regions to find the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {

	// miami https://fl-us-ping.vultr.com/
	// new jersey 1,EWR,New Jersey,NJ,US https://nj-us-ping.vultr.com/
	// chicago 2,ORD,Chicago,IL,US https://il-us-ping.vultr.com
	// dallas https://tx-us-ping.vultr.com
	// seattle 4,SEA,Seattle,WA,US https://wa-us-ping.vultr.com/
	// los angeles https://lax-ca-us-ping.vultr.com/
	// atlanta https://ga-us-ping.vultr.com/
	// silicon valley 12,SJC,Silicon Valley,CA,US https://sjo-ca-us-ping.vultr.com/
	// toronto 22,YTO,Toronto,,CA https://tor-ca-ping.vultr.com

	// amsterdam https://ams-nl-ping.vultr.com/
	// london https://lon-gb-ping.vultr.com/
	// frankfurt https://fra-de-ping.vultr.com/
	// paris https://par-fr-ping.vultr.com/

	// tokyo 25,NRT,Tokyo,,JP https://hnd-jp-ping.vultr.com/
	// seoul 34,Seoul,,KRhttps://sel-kor-ping.vultr.com/
	// singapore 40,SGP, Singapore,,SG https://sgp-ping.vultr.com/

	// sydney https://syd-au-ping.vultr.com/

	regions, err := m.api.Region.List(context.Background())
	if err != nil {
		return "", err
	}

	for _, region := range regions {
		fmt.Println(region.Name, region.Country, region.State, region.RegionCode, region.RegionID)
	}
	return "", nil

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

	servers, err := m.api.Server.List(context.Background())
	for _, server := range servers {
		fmt.Println(server.InstanceID, server.Status, server.Location, server.MainIP)
	}
	return vmInfo, nil
}

// DeleteVM will delete a given VM
func (m *Manager) DeleteVM(ID int) error {

	instanceID := string(ID)
	err := m.api.Server.Delete(context.Background(), instanceID)
	if err != nil {
		return err
	}

	return nil
}
