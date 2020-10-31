package vultr

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/eezhee/eezhee/pkg/core"
	"github.com/go-ping/ping"
	"github.com/vultr/govultr"
)

const maxPingTime = 2 * time.Second // assumes there will be atleast one region closer than this

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

// TODO make more generic
//      should be able to pass array of IPs/hostnames and get sorted array back
//      not just for regions.  can be for hosts (or whatever)

type regionPingTime struct {
	id      string // region ID
	name    string // region name/code
	address string // ip address or hostname (in region) that we can use for ping test
	time    int    // ping time for given ip address (in msec)
}

// getPingTime will do a ping test to the given host / ip address
func getPingTime(region regionPingTime, ch chan regionPingTime) {

	pinger, err := ping.NewPinger(region.address)
	if err == nil {

		// set ping parameters
		pinger.Count = 3
		pinger.Timeout = time.Millisecond * maxPingTime // milliseconds

		// do the ping test
		err = pinger.Run() // blocks until finished
		if err == nil {
			// get results
			stats := pinger.Statistics() // get send/receive/rtt stats

			// save results
			pingTime := stats.AvgRtt.Milliseconds()
			region.time = int(pingTime)
		}
	}

	if err != nil {
		// log the error
		fmt.Println(err)
	}

	// pass results back to caller
	// note: need to pass something back whether worked or not
	// callers waits until gets all results
	ch <- region

	return
}

// SelectClosestRegion will ping all regions and return the ID of the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {

	regionIPs := []regionPingTime{
		{"3", "DFW", "tx-us-ping.vultr.com", 0},
		{"5", "LAX", "lax-ca-us-ping.vultr.com", 0},
		{"39", "MIA", "fl-us-ping.vultr.com", 0},
		{"12", "SJC", "sjo-ca-us-ping.vultr.com", 0},
		{"2", "ORD", "il-us-ping.vultr.com", 0},
		{"4", "SEA", "wa-us-ping.vultr.com", 0},
		{"1", "EWR", "nj-us-ping.vultr.com", 0},
		{"6", "ATL", "ga-us-ping.vultr.com", 0},

		{"22", "TYO", "tor-ca-ping.vultr.com", 0},

		{"24", "CDG", "par-fr-ping.vultr.com", 0},
		{"9", "FRA", "fra-de-ping.vultr.com", 0},
		{"7", "AMS", "ams-nl-ping.vultr.com", 0},
		{"8", "LHR", "lon-gb-ping.vultr.com", 0},

		{"40", "SGP", "sgp-ping.vultr.com", 0},
		{"34", "ICN", "sel-kor-ping.vultr.com", 0},
		{"25", "NRT", "hnd-jp-ping.vultr.com", 0},
		{"19", "SYD", "syd-au-ping.vultr.com", 0},
	}

	numRegions := len(regionIPs)

	// get ping time to each region
	ch := make(chan regionPingTime, numRegions)
	for _, region := range regionIPs {
		go getPingTime(region, ch)
	}

	// reading result until we have them all
	numResults := 0
	lowestPingTime := math.MaxInt64
	for result := range ch {

		numResults++

		// keep track of fastest ping time
		// ignore time of 0 (means ping failed)
		if (result.time > 0) && (result.time < lowestPingTime) {
			closestRegion = result.id
			lowestPingTime = result.time
		}
		// fmt.Println(result.name, result.time)

		// do we have all the results?
		if numResults == numRegions {
			close(ch) // will break the loop
		}
	}

	return closestRegion, nil

	// regions, err := m.api.Region.List(context.Background())
	// if err != nil {
	// 	return "", err
	// }

	// for _, region := range regions {
	// 	fmt.Println(region.Name, region.Country, region.State, region.RegionCode, region.RegionID, region.)
	// }
	// return "", nil

}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID int) (vmInfo core.VMInfo, err error) {

	instanceID := string(vmID)
	_, err = m.api.Server.GetServer(context.Background(), instanceID)
	// server, err := m.api.Server.GetServer(context.Background(), instanceID)
	if err != nil {
		return vmInfo, err
	}

	//Convert to vmInfo format

	return vmInfo, nil
}

// CreateVM will create a new VM
func (m *Manager) CreateVM(name string, image string, size string, region string, sshFingerprint string) (core.VMInfo, error) {
	var vmInfo core.VMInfo

	regionID, _ := strconv.Atoi(region)
	sizeInt, _ := strconv.Atoi(size)
	imageID, _ := strconv.Atoi(image)
	options := govultr.ServerOptions{}
	server, err := m.api.Server.Create(context.Background(), regionID, sizeInt, imageID, &options)
	if err != nil {
		return vmInfo, err
	}
	fmt.Println(server)

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
