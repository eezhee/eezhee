package digitalocean

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

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

// VMInfo has info about a specific VM
type VMInfo struct {
	ID        int         `json:"id"`
	Name      string      `json:"name"`
	Memory    int         `json:"memory"`
	VCPUs     int         `json:"vcpus"`
	Disk      int         `json:"disk"`
	Region    RegionInfo  `json:"region"`
	Status    string      `json:"status"`
	SizeSlug  string      `json:"size_slug"`
	CreatedAt string      `json:"created_at"`
	Image     ImageInfo   `json:"image"`
	Size      SizeInfo    `json:"size"`
	Networks  NetworkInfo `json:"networks"`
	Tags      []string    `json:"tags"`
	VPCUUID   string      `json:"vpc_uuid"`
	// VolumeIDs array of:
}

// RegionInfo has details about a given datacenter
type RegionInfo struct {
	Slug      string   `json:"slug"`
	Name      string   `json:"name"`
	Sizes     []string `json:"sizes"`
	Available bool     `json:"available"`
	Features  []string `json:"features"`
}

// SizeInfo has details about a specific VM size
type SizeInfo struct {
	Slug         string   `json:"slug"`
	Memory       int      `json:"memory"`
	VCPUs        int      `json:"vcpus"`
	Disk         int      `json:"disk"`
	PriceMonthly float32  `json:"price_monthly"`
	PriceHourly  float32  `json:"price_hourly"`
	Regions      []string `json:"regions"`
	Available    bool     `json:"available"`
	Transfer     int      `json:"transfer"`
}

// ImageInfo contains details about a disk image
type ImageInfo struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Distrubution  string   `json:"distribution"`
	Slug          string   `json:"slug"`    // N/A if status is 'retired'
	Public        bool     `json:"public"`  // N/A if status is 'retired'
	Regions       []string `json:"regions"` // N/A if status is 'retired'
	MinDiskSize   int      `json:"min_disk_size"`
	SizeGigabytes float64  `json:"size_gigabytes"`
	CreatedAt     string   `json:"created_at"`
	Description   string   `json:"description"`
	Status        string   `json:"status"`
}

// V4NetworkInfo has details of ipv4 networks
type V4NetworkInfo struct {
	IPAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
	Type      string `json:"type"`
}

// NetworkInfo has info about the networks a VM has
type NetworkInfo struct {
	V4Info []V4NetworkInfo `json:"v4"`
}

// Manager handles interactions with DigitalOcean API
type Manager struct {
}

// NewManager creates a manage object & inits it
func NewManager() (m *Manager) {
	manager := new(Manager)
	return manager
}

// CheckRequirements makes sure necessary things installed to talk to DO
func (m *Manager) CheckRequirements() (bool, error) {

	// is doctl installed
	cmd := exec.Command("which", "doctl")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.New("doctl is not installed")
	}

	// has user authenticated
	// should list at least one item, normally is `default` context
	cmd = exec.Command("doctl", "auth", "list")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return false, errors.New("doctl is not logged in.  Use 'doctl auth init'")
	}

	return true, nil
}

// SSHKeys has details of a ssh key in user's DO account
type SSHKeys struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
}

// CheckSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) CheckSSHKeyUploaded(fingerprint string) (bool, error) {

	var sshKeys []SSHKeys

	// get list of sshkeys DO knows about
	cmd := exec.Command("doctl", "compute", "ssh-key", "list", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	err := json.Unmarshal([]byte(stdoutStderr), &sshKeys)
	if err != nil {
		return false, err
	}

	// go through each key and see if it matches what is on this machine
	for i := 0; i < len(sshKeys); i++ {
		if strings.Compare(fingerprint, sshKeys[i].Fingerprint) == 0 {
			// fmt.Println("found ssh key")
			return true, nil
		}
	}

	return false, errors.New("ssh key not available.  you need to upload it to DO using the web console")
}

type regionPingTimes struct {
	name      string
	ipAddress string
}

// type sampleIPAddress struct {
// 	region    string
// 	ipAddress string
// }

func getPingTime(ipAddress string) (pingTime int64, err error) {

	pinger, err := ping.NewPinger(ipAddress)
	pinger.Timeout = time.Millisecond * maxPingTime // milliseconds
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	pinger.Count = 5
	err := pinger.Run() // blocks until finished
	if err != nil {
		return 0, err
	}
	stats := pinger.Statistics() // get send/receive/rtt stats

	pingTime = stats.AvgRtt.Milliseconds()

	return pingTime, nil
}

// SelectClosestRegion will check all DO regions to find the closest
func (m *Manager) SelectClosestRegion() (string, error) {

	sampleIPs := []regionPingTimes{
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
	var bestRegion = "nyc1" // default to nyc

	// get ping time to each region
	// to see which is the closest
	var lowestPingTime = maxPingTime
	for _, region := range sampleIPs {
		pingTime, err := getPingTime(region.ipAddress)
		if err != nil {
			return "", err
		}
		// fmt.Println(region.name, ": ", pingTime, "mSec")

		// is this datacenter closer than others we've seen so far
		if int(pingTime) < lowestPingTime {
			bestRegion = region.name
			lowestPingTime = int(pingTime)
		}
	}

	return bestRegion, nil
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID int) ([]VMInfo, error) {

	var vmInfo []VMInfo

	// get the latest VM info.  see if status active now
	getCmd := exec.Command("doctl", "compute", "droplet", "get", strconv.Itoa(vmID), "-o", "json")
	stdoutStderr, err := getCmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(stdoutStderr))
		return vmInfo, err
	}

	// parse the json output
	err := json.Unmarshal([]byte(stdoutStderr), &vmInfo)
	if err != nil {
		return vmInfo, err
	}

	return vmInfo, nil
}

// GetPublicIP for the VM
func (v *VMInfo) GetPublicIP() (publicIP string, err error) {

	// go through all network and find which one is public
	numNetworks := len(v.Networks.V4Info)

	for i := 0; i < numNetworks; i++ {

		networkType := v.Networks.V4Info[i].Type

		if strings.Compare(networkType, "public") == 0 {
			publicIP := v.Networks.V4Info[i].IPAddress
			return publicIP, nil
		}

	}

	// did not find public IP
	return publicIP, errors.New("VM does not have public IP")
}

// CreateVM will create a new VM
func (m *Manager) CreateVM(name string, image string, size string, region string, sshFingerprint string) ([]VMInfo, error) {

	var vmInfo []VMInfo

	// create the vm
	cmd := exec.Command("doctl", "compute", "droplet", "create", name,
		"--image", image, "--size", size, "--region", region, "--ssh-keys", sshFingerprint,
		"--tag-name", "eezhee", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(stdoutStderr))
		fmt.Println(err)
		return vmInfo, err
	}

	// parse the json output
	err := json.Unmarshal([]byte(stdoutStderr), &vmInfo)
	if err != nil {
		return vmInfo, err
	}

	return vmInfo, nil
}

// ListVMs will return a list of all VMs created by eezhee
func (m *Manager) ListVMs() (vmInfo []VMInfo, err error) {

	// get a list of VMs running on DO
	cmd := exec.Command("doctl", "compute", "droplet", "list", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		// fmt.Println(err)
		return nil, err
	}

	// parse the json output
	var info []VMInfo
	json.Unmarshal([]byte(stdoutStderr), &info)

	// go through all VMs and look for VMs that are tagged with 'eezhee'
	for i := range info {
		if len(info[i].Tags) > 0 {
			for _, tag := range info[i].Tags {
				if strings.Compare(tag, "eezhee") == 0 {
					// we created this VM
					vmInfo = append(vmInfo, info[i])
					// fmt.Println(vmInfo[i].Name, " (", vmInfo[i].ID, ")  status:", vmInfo[i].Status, " created at:", vmInfo[i].CreatedAt)
				}
			}
		}
	}

	return vmInfo, nil
}

// DeleteVM will delete a given VM
func (m *Manager) DeleteVM(ID int) error {

	// doctl compute droplet delete id
	cmd := exec.Command("doctl", "compute", "droplet", "delete", strconv.Itoa(ID), "--force", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(stdoutStderr))
		fmt.Println(err)
		return err
	}

	fmt.Println(string(stdoutStderr))

	return nil
}

// ComputeDropletCreate will create a new VM
func ComputeDropletCreate() (vmInfo VMInfo, err error) {
	return vmInfo, err
}
