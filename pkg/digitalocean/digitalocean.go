package digitalocean

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

// ComputeDropletCreate will create a new VM
func ComputeDropletCreate() (vmInfo VMInfo, err error) {
	return vmInfo, err
}

// ComputeDropletList will return info on VMs created by Eezhee
func ComputeDropletList() error {
	return nil
}
