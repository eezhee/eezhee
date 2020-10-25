package linode

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eezhee/eezhee/pkg/core"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

// Manager handles interactions with DigitalOcean API
type Manager struct {
	APIToken string
	api      linodego.Client
}

// NewManager creates a manage object & inits it
func NewManager(providerAPIToken string) (m *Manager) {

	if len(providerAPIToken) == 0 {
		fmt.Println("no linode api token set")
		return nil
	}

	manager := new(Manager)
	manager.APIToken = providerAPIToken
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: manager.APIToken})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	manager.api = linodego.NewClient(oauth2Client)

	return manager
}

// IsSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) IsSSHKeyUploaded(fingerprint string) (bool, error) {

	// ctx := context.TODO()
	return true, nil
}

// SelectClosestRegion will check all DO regions to find the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {
	return "", nil
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID int) (vmInfo core.VMInfo, err error) {

	// ctx := context.TODO()
	kernels, err := m.api.ListKernels(context.Background(), nil)
	fmt.Println(len(kernels))

	return vmInfo, nil
}
