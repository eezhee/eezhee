package vultr

import (
	"context"
	"strconv"
	"strings"

	"github.com/eezhee/eezhee/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/vultr/govultr/v2"
	"golang.org/x/crypto/ssh"
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
	{ID: "3", Address: "tx-us-ping.vultr.com"},
	{ID: "5", Address: "lax-ca-us-ping.vultr.com"},
	{ID: "39", Address: "fl-us-ping.vultr.com"},
	{ID: "12", Address: "sjo-ca-us-ping.vultr.com"},
	{ID: "2", Address: "il-us-ping.vultr.com"},
	{ID: "4", Address: "wa-us-ping.vultr.com"},
	{ID: "1", Address: "nj-us-ping.vultr.com"},
	{ID: "6", Address: "ga-us-ping.vultr.com"},
	{ID: "22", Address: "tor-ca-ping.vultr.com"},
	{ID: "24", Address: "par-fr-ping.vultr.com"},
	{ID: "9", Address: "fra-de-ping.vultr.com"},
	{ID: "7", Address: "ams-nl-ping.vultr.com"},
	{ID: "8", Address: "lon-gb-ping.vultr.com"},
	{ID: "40", Address: "sgp-ping.vultr.com"},
	{ID: "34", Address: "sel-kor-ping.vultr.com"},
	{ID: "25", Address: "hnd-jp-ping.vultr.com"},
	{ID: "19", Address: "syd-au-ping.vultr.com"},
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

	manager.api = govultr.NewClient(nil, manager.APIToken)

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
	keys, err := m.api.SSHKey.List(context.Background())
	if err != nil {
		return "", err
	}

	// see if desired key is on the list
	haveKey := false
	for _, key := range keys {

		// log.Debug(key.Name)

		sshKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key.Key))
		if err != nil {
			continue
		}
		fingerprint := ssh.FingerprintLegacyMD5(sshKey)
		// check fingerprint and see if we have a match
		if fingerprint == desiredSSHKey.Fingerprint() {
			haveKey = true
			keyID = key.SSHKeyID
			break
		}
	}

	// we need to add the key
	if !haveKey {
		key, err := m.api.SSHKey.Create(context.Background(), "athir-eezhee", desiredSSHKey.GetPublicKey())
		if err != nil {
			return "", err
		}
		keyID = key.SSHKeyID
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
func (m *Manager) GetVMInfo(vmID int) (vmInfo core.VMInfo, err error) {

	instanceID := strconv.Itoa(vmID)
	server, err := m.api.Server.GetServer(context.Background(), instanceID)
	if err != nil {
		return vmInfo, err
	}

	//Convert info to our format
	vmInfo.ID, _ = strconv.Atoi(server.InstanceID)
	vmInfo.Name = server.Label
	vmInfo.Region = core.RegionInfo{
		Name: server.Location,
		Slug: server.RegionID,
	}
	//TODO set RAM, Disk, VPSCpus, Cost (both in size and in vminfo)
	// issue is these are not standard formats
	vmInfo.Size = core.SizeInfo{
		Slug: server.PlanID,
	}
	imageID, _ := strconv.Atoi(server.OsID)
	vmInfo.Image = core.ImageInfo{
		ID:   imageID,
		Name: server.Os,
	}
	vmInfo.CreatedAt = server.Created
	vmInfo.Tags = append(vmInfo.Tags, server.Tag)

	// only status that needs to be standardized is final one that server is up
	// at vultr that is "ok"
	if strings.Compare(server.ServerState, "ok") == 0 {
		vmInfo.Status = "running"
	} else {
		if strings.Compare(server.Status, "pending") == 0 { // TODO: powerstatus = 'stopped' show that rather than 'locked'
			// serverstatus will be 'none' so use status instead
			vmInfo.Status = server.Status
		} else {
			vmInfo.Status = server.ServerState
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

	regionID, _ := strconv.Atoi(region)
	sizeInt, _ := strconv.Atoi(size)
	imageID, _ := strconv.Atoi(image)
	options := govultr.ServerOptions{
		Label:     name,
		SSHKeyIDs: keyIDs,
		Tag:       "eezhee",
	}
	server, err := m.api.Server.Create(context.Background(), regionID, sizeInt, imageID, &options)
	if err != nil {
		return vmInfo, err
	}
	log.Info("vm ", server.InstanceID, " created")

	// transfer data to vmInfo
	vmInfo.ID, err = strconv.Atoi(server.InstanceID)
	if err != nil {
		return vmInfo, err
	}

	return vmInfo, nil
}

// ListVMs will return a list of all VMs created by eezhee
func (m *Manager) ListVMs() (vmInfo []core.VMInfo, err error) {

	instances, err := m.api.Server.List(context.Background())
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

		log.Debug(instance.InstanceID, " ", instance.Status, " ", instance.Location, " ", instance.MainIP)
	}
	return vmInfo, nil
}

// convertVMInfoToGenericFormat cloud vendor info into our generic format
func convertVMInfoToGenericFormat(instance govultr.Instance) (core.VMInfo, error) {

	var vmInfo core.VMInfo

	vmInfo.ID, _ = strconv.Atoi(instance.InstanceID)

	vmInfo.Name = instance.Label

	vmInfo.Memory, _ = strconv.Atoi(instance.RAM)
	vmInfo.VCPUs, _ = strconv.Atoi(instance.VPSCpus)
	vmInfo.Disk = instance.Disk

	vmInfo.Region = core.RegionInfo{Name: instance.Region}
	vmInfo.Status = string(instance.Status)

	vmInfo.CreatedAt = instance.Created.String()

	vmInfo.Image = core.ImageInfo{
		// ID: instance.Image,   	// int vs string
		Name: instance.Image,
	}

	vmInfo.Size = core.SizeInfo{
		Slug: instance.Type,
	}
	vmInfo.Networks = core.NetworkInfo{
		V4Info: []core.V4NetworkInfo{},
		V6Info: []core.V6NetworkInfo{},
	}

	v4NetworkInfo := core.V4NetworkInfo{
		IPAddress: instance.IPv4[0].String(),
	}
	vmInfo.Networks.V4Info = append(vmInfo.Networks.V4Info, v4NetworkInfo)

	v6NetworkInfo := core.V6NetworkInfo{
		IPAddress: instance.IPv6,
	}
	vmInfo.Networks.V6Info = append(vmInfo.Networks.V6Info, v6NetworkInfo)

	vmInfo.Tags = instance.Tags

	return vmInfo, nil
}

// DeleteVM will delete a given VM
func (m *Manager) DeleteVM(ID int) error {

	// NOTE: if current status is 'pending' then can't delete (yet)
	//       need to wait until build completed first

	instanceID := strconv.Itoa(ID)
	err := m.api.Server.Delete(context.Background(), instanceID)
	if err != nil {
		return err
	}

	return nil
}
