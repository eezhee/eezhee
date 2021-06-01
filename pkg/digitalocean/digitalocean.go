package digitalocean

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/eezhee/eezhee/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//go:embed digitalocean-mappings.json
var DigitalOceanMappingsJSON []byte

// ip addresses to use to find closest region
var regionIPs = []core.IPPingTime{
	{ID: "ams2", Address: "206.189.240.1"},
	{ID: "blr1", Address: "143.110.180.2"},
	{ID: "fra1", Address: "138.68.109.1"},
	{ID: "lon1", Address: "209.97.176.1"},
	{ID: "nyc1", Address: "192.241.251.1"},
	{ID: "sfo1", Address: "198.199.113.1"},
	{ID: "sgp1", Address: "209.97.160.1"},
	{ID: "tor1", Address: "68.183.194.1"},
}

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

// TODO: the way to manage sizes is on a 3 (or more dimensional plane).  User decides what they want to increase and we figure out the right VM upgrade

// Manager handles interactions with DigitalOcean API
type Manager struct {
	APIToken string
	api      *godo.Client
}

// NewManager creates a manage object & inits it
func NewManager(providerAPIToken string) (core.VMManager, error) {

	manager := new(Manager)

	// make sure we have an api token
	if len(providerAPIToken) == 0 {
		// check places provider CLI tools store token
		providerAPIToken := manager.FindAuthToken()
		if len(providerAPIToken) == 0 {
			return manager, errors.New("no digitalocean api token set")
		}
		// ok we found a token
	}

	manager.APIToken = providerAPIToken
	manager.api = godo.NewFromToken(manager.APIToken)

	return manager, nil
}

// FindAuthToken will check common place for digitalocean api key
func (m *Manager) FindAuthToken() string {

	accessToken := ""

	// doctl calls os.UserConfigDir() to figure out where to save config file

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		log.Debug("could not get home dir when looking for digitalocean config:", err)
		return ""
	}

	cfgPath := filepath.Join(cfgDir, "doctl")
	cfgFile := filepath.Join(cfgPath, "config.yaml")

	config := viper.New()
	config.SetEnvPrefix("DIGITALOCEAN")
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	config.SetConfigType("yaml")
	config.SetConfigFile(cfgFile)

	if _, err := os.Stat(cfgFile); err == nil {
		if err := config.ReadInConfig(); err != nil {
			log.Debug("could not read digitalocean config:", err)
		}

		accessToken = config.GetString("access-token")
	}

	return accessToken
}

// GetRegions will return all the regions that Eezhee can use
func (m *Manager) GetRegions() ([]string, error) {

	list := []string{}

	// get mappings from DO format to Eezhee format
	mappings, err := core.ParseProviderMappings(DigitalOceanMappingsJSON)
	if err != nil {
		return list, err
	}

	// go through regions
	ctx := context.Background()
	for {

		// get regions
		page := 1
		regions, resp, err := m.api.Regions.List(ctx, &godo.ListOptions{Page: page})
		if err != nil {
			return list, err
		}

		// go through regions and compare to our mapping
		for _, region := range regions {

			// is region supported
			mapping, ok := mappings.Regions[region.Slug]
			if ok {

				// convert from DO_region to EZ_region
				// output both our name(s) and DO name
				// do we need resp.  has links and meta fields
				var alternates = make([]string, 0)
				alternates = append(alternates, mapping.Country)
				if len(mapping.Region) > 0 {
					alternates = append(alternates, mapping.Country+"-"+mapping.Region)
				}
				if len(mapping.State) > 0 {
					alternates = append(alternates, mapping.Country+"-"+mapping.State)
				}
				fmt.Printf("%s: %s (%s)\n", mapping.City, alternates, region.Slug)

			}
		}

		// see if there are more pages
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		} else {
			page = page + 1
		}

	}

	return list, nil
}

// GetSizes will return all the VM sizes that DO has (and that Eezhee can use)
func (m *Manager) GetVMSizes() ([]string, error) {

	list := []string{}

	// get mappings from DO format to Eezhee format
	mappings, err := core.ParseProviderMappings(DigitalOceanMappingsJSON)
	if err != nil {
		return list, err
	}

	// go through regions
	ctx := context.Background()
	page := 1
	for {

		// get regions
		sizes, resp, err := m.api.Sizes.List(ctx, &godo.ListOptions{Page: page})
		if err != nil {
			return list, err
		}

		// go through regions and compare to our mapping
		for _, size := range sizes {

			// is region supported
			name, ok := mappings.VMSizes[size.Slug]
			if ok {
				fmt.Printf("%s: disk: %vGB xfer: %vTB (%s) \n", name, size.Disk, size.Transfer, size.Slug)
			}
		}

		// see if there are more pages
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		} else {
			page = page + 1
		}

	}

	return list, nil
}

// IsSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) IsSSHKeyUploaded(desiredSSHKey core.SSHKey) (string, error) {

	ctx := context.TODO()

	// get list of sshkeys DO knows about
	sshKeys, _, err := m.api.Keys.List(ctx, nil)
	if err != nil {
		return "", err
	}

	// go through each key and see if it matches what is on this machine
	for _, sshKey := range sshKeys {
		if strings.Compare(desiredSSHKey.Fingerprint(), sshKey.Fingerprint) == 0 {
			keyID := strconv.Itoa(sshKey.ID)
			return keyID, nil
		}
	}

	return "", errors.New("ssh key not available on digitalocean")
}

// UploadSSHKey will upload given key to the provider
func (m *Manager) UploadSSHKey(keyName string, sshKey core.SSHKey) (keyID string, err error) {

	// if provider account shared with  more than one person, key name needs to be unique
	// let's add first few characters  fingerprint
	fingerprint := sshKey.Fingerprint()
	keyName = keyName + "-" + fingerprint[0:8]

	createRequest := &godo.KeyCreateRequest{
		Name:      keyName,
		PublicKey: sshKey.GetPublicKey(),
	}

	ctx := context.Background()
	key, _, err := m.api.Keys.Create(ctx, createRequest)
	if err != nil {
		msg := fmt.Sprintf("could not upload ssh key: %s", err)
		return "", errors.New(msg)
	}

	id := strconv.Itoa(key.ID)
	return id, nil
}

// SelectClosestRegion will check all DO regions to find the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {
	closestRegion, err = core.GetPingTimes(regionIPs)
	// note regionsIPs is now filled with ping times
	return closestRegion, err
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID string) (vmInfo core.VMInfo, err error) {

	// get the latest VM info.  see if status active now
	ctx := context.TODO()
	instanceID, _ := strconv.Atoi(vmID)
	droplet, _, err := m.api.Droplets.Get(ctx, instanceID)
	if err != nil {
		log.Error(err)
		return vmInfo, err
	}

	// need to convert info from digitalocean format to our format
	vmInfo, _ = convertVMInfoToGenericFormat(*droplet)

	return vmInfo, nil
}

// CreateVM will create a new VM
func (m *Manager) CreateVM(name string, image string, size string, region string, sshKey core.SSHKey) (core.VMInfo, error) {

	var vmInfo core.VMInfo

	createRequest := &godo.DropletCreateRequest{
		Name:   name,
		Region: region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{{Fingerprint: sshKey.Fingerprint()}},
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

	log.Debug("account has ", len(droplets), " VMs")

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
func (m *Manager) DeleteVM(ID string) error {

	ctx := context.TODO()

	instanceID, _ := strconv.Atoi(ID)
	_, err := m.api.Droplets.Delete(ctx, instanceID)
	if err != nil {
		return err
	}

	log.Debug("vm ", ID, " deleted")

	return nil
}

// convert digitalocean droplet info into our generic format
func convertVMInfoToGenericFormat(dropletInfo godo.Droplet) (core.VMInfo, error) {

	var vmInfo core.VMInfo

	vmInfo.ID = strconv.Itoa(dropletInfo.ID)
	vmInfo.Name = dropletInfo.Name
	vmInfo.Memory = dropletInfo.Memory / 1024
	vmInfo.VCPUs = dropletInfo.Vcpus
	vmInfo.Disk = dropletInfo.Disk
	vmInfo.Region = core.RegionInfo{
		Name: dropletInfo.Region.Name,
		Slug: dropletInfo.Region.Slug,
		// Features: dropletInfo.Region.Features,
	}
	// need to convert final status to standard format
	if dropletInfo.Status == "active" {
		vmInfo.Status = "running"
	} else {
		vmInfo.Status = dropletInfo.Status
	}
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
	vmInfo.SizeSlug = dropletInfo.Size.Slug
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
