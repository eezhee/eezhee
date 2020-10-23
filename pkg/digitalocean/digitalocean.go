package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/eezhee/eezhee/pkg/core"
	"github.com/go-ping/ping"
)

const maxPingTime = 750

// datacenters: ams2","ams3","blr1","fra1","lon1","nyc1","nyc2","nyc3","sfo1","sfo2","sfo3","sgp1","tor1"
// sizes:
//        "512mb","1gb","2gb","4gb","8gb","16gb","32gb","48gb","64gb",
//        "s-1vcpu-1gb","s-1vcpu-2gb","s-3vcpu-1gb","s-2vcpu-2gb","s-1vcpu-3gb","s-2vcpu-4gb","s-4vcpu-8gb","s-8vcpu-16gb","s-6vcpu-16gb",
//			  "s-8vcpu-32gb","s-12vcpu-48gb","s-16vcpu-64gb","s-20vcpu-96gb","s-24vcpu-128gb","s-32vcpu-192gb",
//        "m-16gb","m-32gb","m-64gb","m-128gb","m-224gb",
//        "m-1vcpu-8gb","m-2vcpu-16gb","m-4vcpu-32gb","m-8vcpu-64gb","m-16vcpu-128gb","m-24vcpu-192gb","m-32vcpu-256gb",
//				"m3-2vcpu-16gb","m3-4vcpu-32gb","m3-8vcpu-64gb","m3-16vcpu-128gb","m3-24vcpu-192gb","m3-32vcpu-256gb",
//				"m6-2vcpu-16gb","m6-4vcpu-32gb","m6-8vcpu-64gb","m6-16vcpu-128gb","m6-24vcpu-192gb","m6-32vcpu-256gb"
//        "c-2","c2-2vcpu-4gb","c-4","c2-4vpcu-8gb","c-8","c2-8vpcu-16gb","c-16","c2-16vcpu-32gb","c-32","c2-32vpcu-64gb",
// images:  no need as we always want to be on the same plain Ubuntu box

// TODO: the way to manage sizes is on a 3 (or more dimensional plane).  User decides what they want to increase and we figure out the right VM upgrade

// Manager handles interactions with DigitalOcean API
type Manager struct {
	APIToken string
	api      *godo.Client
}

// NewManager creates a manage object & inits it
func NewManager(providerAPIToken string) (m *Manager) {

	if len(providerAPIToken) == 0 {
		fmt.Println("no digitalocean api token set")
		return nil
	}

	manager := new(Manager)
	manager.APIToken = providerAPIToken
	manager.api = godo.NewFromToken(manager.APIToken)

	return manager
}

// IsSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) IsSSHKeyUploaded(fingerprint string) (bool, error) {

	ctx := context.TODO()

	// get list of sshkeys DO knows about
	sshKeys, _, err := m.api.Keys.List(ctx, nil)
	if err != nil {
		return false, err
	}

	// go through each key and see if it matches what is on this machine
	for _, sshKey := range sshKeys {
		if strings.Compare(fingerprint, sshKey.Fingerprint) == 0 {
			return true, nil
		}
	}

	return false, errors.New("ssh key not available on digitalocean")
}

type regionPingTimes struct {
	name      string
	ipAddress string
}

func getPingTime(ipAddress string) (pingTime int64, err error) {

	pinger, err := ping.NewPinger(ipAddress)
	// pinger.Timeout = time.Millisecond * maxPingTime // milliseconds
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	pinger.Count = 3
	err = pinger.Run() // blocks until finished
	if err != nil {
		return 0, err
	}
	stats := pinger.Statistics() // get send/receive/rtt stats

	pingTime = stats.AvgRtt.Milliseconds()

	return pingTime, nil
}

// SelectClosestRegion will check all DO regions to find the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {

	regionIPs := []regionPingTimes{
		{"ams2", "206.189.240.1"},
		{"blr1", "143.110.180.2"},
		{"fra1", "138.68.109.1"},
		{"lon1", "209.97.176.1"},
		{"nyc1", "192.241.251.1"},
		{"sfo1", "198.199.113.1"},
		{"sgp1", "209.97.160.1"},
		{"tor1", "68.183.194.1"},
	}

	// default to NYC
	closestRegion = "nyc1"

	// get ping time to each region
	// to see which is the closest
	var lowestPingTime = maxPingTime
	for _, region := range regionIPs {
		pingTime, err := getPingTime(region.ipAddress)
		if err != nil {
			return "", err
		}
		// fmt.Println(region.name, ": ", pingTime, "mSec")

		// is this datacenter closer than others we've seen so far
		if int(pingTime) < lowestPingTime {
			closestRegion = region.name
			lowestPingTime = int(pingTime)
		}
	}

	return closestRegion, nil
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID int) (vmInfo core.VMInfo, err error) {

	// get the latest VM info.  see if status active now
	ctx := context.TODO()
	droplet, _, err := m.api.Droplets.Get(ctx, vmID)
	if err != nil {
		fmt.Println(err)
		return vmInfo, err
	}

	// need to convert info from digitalocean format to our format
	vmInfo, _ = convertVMInfoToGenericFormat(*droplet)

	return vmInfo, nil
}

// CreateVM will create a new VM
func (m *Manager) CreateVM(name string, image string, size string, region string, sshFingerprint string) (core.VMInfo, error) {

	var vmInfo core.VMInfo

	createRequest := &godo.DropletCreateRequest{
		Name:   name,
		Region: region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{{Fingerprint: sshFingerprint}},
		// Volumes: []godo.DropletCreateVolume{
		// 	{Name: "hello-im-a-volume"},
		// 	{ID: "hello-im-another-volume"},
		// 	{Name: "hello-im-still-a-volume", ID: "should be ignored due to Name"},
		// },
		// VPCUUID: "880b7f98-f062-404d-b33c-458d545696f6",
		Tags: []string{"eezhee"},
	}
	ctx := context.TODO()

	newDroplet, _, err := m.api.Droplets.Create(ctx, createRequest)
	if err != nil {
		return vmInfo, err
	}

	vmInfo, _ = convertVMInfoToGenericFormat(*newDroplet)

	return vmInfo, nil
}

// ListVMs will return a list of all VMs created by eezhee
func (m *Manager) ListVMs() (vmInfo []core.VMInfo, err error) {

	ctx := context.TODO()

	// get a list of VMs running on DO
	options := godo.ListOptions{}
	droplets, _, err := m.api.Droplets.List(ctx, &options)
	if err != nil {
		return nil, err
	}

	// go through all VMs and look for VMs that are tagged with 'eezhee'
	for i := range droplets {
		if len(droplets[i].Tags) > 0 {
			for _, tag := range droplets[i].Tags {
				if strings.Compare(tag, "eezhee") == 0 {
					// we created this VM
					info, _ := convertVMInfoToGenericFormat(droplets[i])
					vmInfo = append(vmInfo, info)
				}
			}
		}
	}

	return vmInfo, nil
}

// DeleteVM will delete a given VM
func (m *Manager) DeleteVM(ID int) error {

	ctx := context.TODO()

	_, err := m.api.Droplets.Delete(ctx, ID)
	if err != nil {
		return err
	}

	return nil
}

// convert digitalocean droplet info into our generic format
func convertVMInfoToGenericFormat(dropletInfo godo.Droplet) (core.VMInfo, error) {

	var vmInfo core.VMInfo

	vmInfo.ID = dropletInfo.ID
	vmInfo.Name = dropletInfo.Name
	vmInfo.Memory = dropletInfo.Memory
	vmInfo.VCPUs = dropletInfo.Vcpus
	vmInfo.Disk = dropletInfo.Disk
	vmInfo.Region = core.RegionInfo{
		Name:     dropletInfo.Region.Name,
		Slug:     dropletInfo.Region.Slug,
		Features: dropletInfo.Region.Features,
	}
	vmInfo.Status = dropletInfo.Status
	// vmInfo.SizeSlug = dropletInfo.SizeSlug
	vmInfo.CreatedAt = dropletInfo.Created
	vmInfo.Image = core.ImageInfo{
		ID:           dropletInfo.Image.ID,
		Name:         dropletInfo.Image.Name,
		Description:  dropletInfo.Image.Description,
		Type:         dropletInfo.Image.Type,
		Distrubution: dropletInfo.Image.Distribution,
		Slug:         dropletInfo.Image.Slug,
		CreatedAt:    dropletInfo.Image.Created,
	}
	vmInfo.Size = core.SizeInfo{
		Slug: dropletInfo.Size.Slug,
	}
	vmInfo.Networks = core.NetworkInfo{
		V4Info: []core.V4NetworkInfo{},
		V6Info: []core.V6NetworkInfo{},
	}

	for _, ipv4Info := range dropletInfo.Networks.V4 {
		v4NetworkInfo := core.V4NetworkInfo{
			IPAddress: ipv4Info.IPAddress,
			Netmask:   ipv4Info.Netmask,
			Gateway:   ipv4Info.Gateway,
			Type:      ipv4Info.Type,
		}
		vmInfo.Networks.V4Info = append(vmInfo.Networks.V4Info, v4NetworkInfo)
	}
	for _, ipv6Info := range dropletInfo.Networks.V6 {
		v6NetworkInfo := core.V6NetworkInfo{
			IPAddress: ipv6Info.IPAddress,
			Netmask:   ipv6Info.Netmask,
			Gateway:   ipv6Info.Gateway,
			Type:      ipv6Info.Type,
		}
		vmInfo.Networks.V6Info = append(vmInfo.Networks.V6Info, v6NetworkInfo)
	}

	vmInfo.VPCUUID = dropletInfo.VPCUUID
	vmInfo.Tags = dropletInfo.Tags

	return vmInfo, nil
}
