package vultr

import (
	"context"
	"strconv"
	"strings"

	"github.com/eezhee/eezhee/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/vultr/govultr/v2"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

// Plan details about each VM plan
type Plan struct {
	ID        string
	Name      string
	VCPUs     int
	RAM       int
	Disk      int
	Bandwidth int //in GB
	Price     float32
}

// plans are ordered by price, which normally correlates with larger CPU,RAM & disk
// TODO: more dynamic way would be to get list and sort by price
// var planOrder []int = []int{201, 202, 203, 204, 205, 206, 207, 208}

var regionIPs = []core.IPPingTime{
	{ID: "dfw", Address: "tx-us-ping.vultr.com"},
	{ID: "lax", Address: "lax-ca-us-ping.vultr.com"},
	{ID: "mia", Address: "fl-us-ping.vultr.com"},
	{ID: "sjc", Address: "sjo-ca-us-ping.vultr.com"},
	{ID: "ord", Address: "il-us-ping.vultr.com"},
	{ID: "sea", Address: "wa-us-ping.vultr.com"},
	{ID: "ewr", Address: "nj-us-ping.vultr.com"},
	{ID: "atl", Address: "ga-us-ping.vultr.com"},
	{ID: "yto", Address: "tor-ca-ping.vultr.com"},
	{ID: "cdg", Address: "par-fr-ping.vultr.com"},
	{ID: "fra", Address: "fra-de-ping.vultr.com"},
	{ID: "ams", Address: "ams-nl-ping.vultr.com"},
	{ID: "lhr", Address: "lon-gb-ping.vultr.com"},
	{ID: "sgp", Address: "sgp-ping.vultr.com"},
	{ID: "icn", Address: "sel-kor-ping.vultr.com"},
	{ID: "nrt", Address: "hnd-jp-ping.vultr.com"},
	{ID: "syd", Address: "syd-au-ping.vultr.com"},
}

// Manager controls access to AWS
type Manager struct {
	APIToken string
	api      *govultr.Client
	// plans    []Plan
}

// NewManager creates a manage object & inits it
func NewManager(providerAPIToken string) (m *Manager) {

	// if user as aws-cli installed, sdk can find ~/.aws/credential file and load keys
	// otherwise, we can create it when user enters details

	if len(providerAPIToken) == 0 {
		log.Error("no vultr api token set")
		return nil
	}

	manager := new(Manager)
	manager.APIToken = providerAPIToken

	config := &oauth2.Config{}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: providerAPIToken})

	manager.api = govultr.NewClient(oauth2.NewClient(ctx, ts))

	return manager
}

// getPlans will get all active plans and sort by price
// func (m *Manager) getCurentPlans() error {

// 	// get plans
// 	plans, err := m.api.Plan.List(context.Background(), "vc2")
// 	if err != nil {
// 		return err
// 	}

// 	for _, plan := range plans {
// 		log.Debug(plan.Name)
// 		// plan 201 is $5, name: "1024 MB RAM,25 GB SSD,1.00 TB BW"
// 	}

// 	// sort in order
// 	// store plan list in object

// 	return nil
// }

// IsSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) IsSSHKeyUploaded(desiredSSHKey core.SSHKey) (keyID string, err error) {

	// get all keys that are on vutrl
	keys, _, err := m.api.SSHKey.List(context.Background(), nil)
	if err != nil {
		return "", err
	}

	// see if desired key is on the list
	haveKey := false
	for _, key := range keys {

		// log.Debug(key.Name)

		sshKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key.SSHKey))
		if err != nil {
			continue
		}
		fingerprint := ssh.FingerprintLegacyMD5(sshKey)
		// check fingerprint and see if we have a match
		if fingerprint == desiredSSHKey.Fingerprint() {
			haveKey = true
			keyID = key.ID
			break
		}
	}

	// we need to add the key
	if !haveKey {

		newKey := &govultr.SSHKeyReq{
			Name:   "athir-eezhee",
			SSHKey: desiredSSHKey.GetPublicKey(),
		}

		key, err := m.api.SSHKey.Create(context.Background(), newKey)
		if err != nil {
			return "", err
		}
		keyID = key.ID
	}

	return keyID, nil
}

// SelectClosestRegion will ping all regions and return the ID of the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {
	closestRegion, err = core.GetPingTimesForArray(regionIPs)
	// note regionsIPs is now filled with ping times
	return closestRegion, err
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID string) (vmInfo core.VMInfo, err error) {

	server, err := m.api.Instance.Get(context.Background(), vmID)
	if err != nil {
		return vmInfo, err
	}

	//Convert info to our format
	vmInfo.ID = server.ID
	vmInfo.Name = server.Label
	vmInfo.Region = core.RegionInfo{
		// Name: server.Region,
		Slug: server.Region,
	}
	//TODO set RAM, Disk, VPSCpus, Cost (both in size and in vminfo)
	// issue is these are not standard formats
	vmInfo.Size = core.SizeInfo{
		Slug: server.Plan,
	}
	imageID, _ := strconv.Atoi(server.Os)
	vmInfo.Image = core.ImageInfo{
		ID:   imageID,
		Name: server.Os,
	}
	vmInfo.CreatedAt = server.DateCreated
	vmInfo.Tags = append(vmInfo.Tags, server.Tag)

	// only status that needs to be standardized is final one that server is up
	// at vultr that is "ok"
	if strings.Compare(server.ServerStatus, "ok") == 0 {
		vmInfo.Status = "running"
	} else {
		if strings.Compare(server.Status, "pending") == 0 { // TODO: powerstatus = 'stopped' show that rather than 'locked'
			// serverstatus will be 'none' so use status instead
			vmInfo.Status = server.Status
		} else {
			vmInfo.Status = server.ServerStatus
		}
	}
	// get public IP address
	vmInfo.Networks.V4Info = append(vmInfo.Networks.V4Info, core.V4NetworkInfo{
		IPAddress: server.MainIP,
		Netmask:   server.NetmaskV4,
		Gateway:   server.GatewayV4,
		Type:      "public",
	})

	return vmInfo, nil
}

// CreateVM will create a new VM
func (m *Manager) CreateVM(name string, image string, size string, region string, sshKey core.SSHKey) (core.VMInfo, error) {
	var vmInfo core.VMInfo

	// find the ssh ID to use
	keyID, err := m.IsSSHKeyUploaded(sshKey)
	if err != nil {
		return vmInfo, err
	}
	keyIDs := []string{keyID}

	imageInt, _ := strconv.Atoi(image)
	options := &govultr.InstanceCreateReq{
		Region:     region,
		Plan:       size,
		OsID:       imageInt,
		Label:      name,
		SSHKeys:    keyIDs,
		EnableIPv6: govultr.BoolToBoolPtr(true),
		Tag:        "eezhee",
	}

	server, err := m.api.Instance.Create(context.Background(), options)
	if err != nil {
		return vmInfo, err
	}
	log.Info("vm ", server.ID, " created")

	// transfer data to vmInfo
	vmInfo.ID = server.ID
	if err != nil {
		return vmInfo, err
	}

	return vmInfo, nil
}

// ListVMs will return a list of all VMs created by eezhee
func (m *Manager) ListVMs() (vmInfo []core.VMInfo, err error) {

	instances, _, err := m.api.Instance.List(context.Background(), nil)
	if err != nil {
		return vmInfo, err
	}
	for _, instance := range instances {
		if len(instance.Tag) > 0 {
			if strings.Compare(instance.Tag, "eezhee") == 0 {
				// we created this VM
				info, _ := convertVMInfoToGenericFormat(instance)
				vmInfo = append(vmInfo, info)
			}
		}

		log.Debug(instance.ID, " ", instance.Status, " ", instance.Region, " ", instance.MainIP)
	}
	return vmInfo, nil
}

// convertVMInfoToGenericFormat cloud vendor info into our generic format
func convertVMInfoToGenericFormat(instance govultr.Instance) (core.VMInfo, error) {

	var vmInfo core.VMInfo

	vmInfo.ID = instance.ID

	vmInfo.Name = instance.Label

	vmInfo.Memory = instance.RAM
	vmInfo.VCPUs = instance.VCPUCount
	vmInfo.Disk = instance.Disk

	vmInfo.Region = core.RegionInfo{Slug: instance.Region}
	vmInfo.Status = string(instance.ServerStatus)

	vmInfo.CreatedAt = instance.DateCreated

	vmInfo.Image = core.ImageInfo{
		ID:   instance.OsID,
		Name: instance.Os,
	}

	vmInfo.Size = core.SizeInfo{
		Slug: instance.Plan,
	}
	vmInfo.SizeSlug = instance.Plan

	vmInfo.Networks = core.NetworkInfo{
		V4Info: []core.V4NetworkInfo{},
		V6Info: []core.V6NetworkInfo{},
	}

	v4NetworkInfo := core.V4NetworkInfo{
		IPAddress: instance.MainIP,
		Gateway:   instance.GatewayV4,
		Netmask:   instance.NetmaskV4,
	}
	vmInfo.Networks.V4Info = append(vmInfo.Networks.V4Info, v4NetworkInfo)

	v6NetworkInfo := core.V6NetworkInfo{
		IPAddress: instance.V6MainIP,
		Gateway:   instance.V6Network,
		Netmask:   instance.V6NetworkSize,
	}
	vmInfo.Networks.V6Info = append(vmInfo.Networks.V6Info, v6NetworkInfo)

	vmInfo.Tags = []string{instance.Tag}

	return vmInfo, nil
}

// DeleteVM will delete a given VM
func (m *Manager) DeleteVM(ID string) error {

	// NOTE: if current status is 'pending' then can't delete (yet)
	//       need to wait until build completed first

	err := m.api.Instance.Delete(context.Background(), ID)
	if err != nil {
		return err
	}

	return nil
}
