package cloud

type InstanceSize struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	ProviderID   string `yaml:"provider_id"`
	ProviderName string `yaml:"provider_name"`
	CPUs         int    `yaml:"cpus"`
	Memory       int    `yaml:"memory"`
	Disk         int    `yaml:"disk"`
	Transfer     int    `yaml:"transfer"`
	Price        int    `yaml:"price"`
}

type InstanceSizes []InstanceSize

type Region struct {
	ID           int    `yaml:"id"`
	Name         string `yaml:"name"`
	ProviderID   int    `yaml:"provider_id"`
	ProviderName string `yaml:"provider_name"`
	// country       string
	// state         string
	// city          string
	// lat           float32
	// long          float32
}

type Regions []Region

// ListRegions returns a list of all regions for a provider
func ListRegions() []Regions {
	var cloudRegions []Regions

	// load the list of regions

	return cloudRegions
}

// ListClostest will return the given number of closest regions
// in order (of which is closest)
func ListClosest(user_lat float32, user_long float32, number int) {

}

// GetRegionDetails will return details about given region
func GetRegionDetails(cloud_id int) {

}

// stories
// 1 given a user's location, return the closest region
// IPToLocation returns country and city/state?
// ListCloset()
// 2 given a user's preferred country (ie US), return the best region
// eezhee regions list - will list regions for each provider (both id and cloud_id)
