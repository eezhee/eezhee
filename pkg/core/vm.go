package core

import (
	"errors"
	"strings"
)

// VMProvider is the interface all cloud provider need to follow
type VMProvider interface {
	Int() bool
	ListVMs()
	CreateVM()
	GetVMInfo()
	DeleteVM()
	UploadSSHKey()
	IsSSHKeyUploaded(fingerprint string) (bool, error)
	SelectClosestRegion()
}

// Regions has details about all the regions a provider supports
type Regions interface {
	GetList() ([]RegionInfo, error)
	GetClosestByPing() ([]RegionInfo, error)
	GetClosestByLatLong(lat float32, long float32) ([]RegionInfo, error)
	GetClosestByCountry(country string) ([]RegionInfo, error)
}

// RegionInfo has details about a given datacenter/region
type RegionInfo struct {
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	Available bool     `json:"available"`
	Country   string   `json:"country"`
	State     string   `json:"State"`
	Sizes     []string `json:"sizes"`
	Features  []string `json:"features"`
}

// VMSizes allows callers to move from one size to another
type VMSizes interface {
	GetList() ([]SizeInfo, error)
	GetInfo(size string) (SizeInfo, error)
	GetLargerSize()
	GetSmallerSize()
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

// V6NetworkInfo has details about ipv6 networks
type V6NetworkInfo struct {
	IPAddress string `json:"ip_address"`
	Netmask   int    `json:"netmask"`
	Gateway   string `json:"gateway"`
	Type      string `json:"type"`
}

// NetworkInfo has info about the networks a VM has
type NetworkInfo struct {
	V4Info []V4NetworkInfo `json:"v4,omitempty"`
	V6Info []V6NetworkInfo `json:"v6,omitempty"`
}

// VMInfo has all details about a given VM
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
	VPCUUID   string      `json:"vpc_uuid"`
	Tags      []string    `json:"tags"`
	// VolumeIDs array of:
}

// SSHKeys has details of a ssh key in user's DO account
type SSHKeys struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
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
