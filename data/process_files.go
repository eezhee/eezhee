// sample code to read in provider data

package main

import (
	"flag"
	"fmt"
	"os"
)

const DATA_PATH = "./raw/"

// var PROVIDERS = []string{"digitalocean", "linode", "vultr"}

// define eezhee sizes
// small, medium, large, xlarge, huge
// 1cpu1gb, 2cpu2gb

type LocationInfo struct {
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
}

type ProviderMappings struct {
	Image   string                  `json:"image"`
	Sizes   map[string]string       `json:"sizes"`
	Regions map[string]LocationInfo `json:"regions"`
}

// define eezhee regions
// us
//   us-east
//     us-southeast, us-northeast,
//   us-central
//     us-southcentral, us-northcentral,
//   us-west,
//     us-northwest, us-southwest
//  ca
//    ca-west
//    ca-central
//    ca-east

// type InstanceSize struct {
// 	ID           string `yaml:"id"`
// 	Name         string `yaml:"name"`
// 	ProviderID   string `yaml:"provider_id"`
// 	ProviderName string `yaml:"provider_name"`
// 	CPUs         int    `yaml:"cpus"`
// 	Memory       int    `yaml:"memory"`
// 	Disk         int    `yaml:"disk"`
// 	Transfer     int    `yaml:"transfer"`
// 	Price        int    `yaml:"price"`
// }

// type InstanceSizes []InstanceSize

// type Region struct {
// 	ID           int    `yaml:"id"`
// 	Name         string `yaml:"name"`
// 	ProviderID   int    `yaml:"provider_id"`
// 	ProviderName string `yaml:"provider_name"`
// 	// country       string
// 	// state         string
// 	// city          string
// 	// lat           float32
// 	// long          float32
// }

// type Regions []Region

// ListRegions returns a list of all regions for a provider
// func ListRegions() []Regions {
// 	var cloudRegions []Regions

// 	// load the list of regions

// 	return cloudRegions
// }

// ListClostest will return the given number of closest regions
// in order (of which is closest)
// func ListClosest(user_lat float32, user_long float32, number int) {

// }

// GetRegionDetails will return details about given region
// func GetRegionDetails(cloud_id int) {

// }

type ProviderImporter interface {
	FindUbuntuImages() bool
	ConvertProviderImageSizes() bool
	ConvertProviderRegions() bool
	ReadMappings() bool
}

// THIS SHOULD GO IN EEZHEE
// readYAMLFiles
// func readYAMLFiles() {
// 	for _, provider := range PROVIDERS {
// 		for _, template := range FILE_TEMPLATES {

// 			filename := "./" + provider + template
// 			fmt.Println("filename:", filename)

// 			// read in the yaml
// 			yamlFile, err := ioutil.ReadFile(filename)
// 			if err != nil {
// 				log.Printf("yamlFile.Get err   #%v ", err)
// 				continue
// 			}

// 			var providerRegions Regions
// 			var providerSizes InstanceSizes

// 			if strings.Contains(filename, "region") {
// 				// regions
// 				err = yaml.Unmarshal(yamlFile, providerRegions)
// 				if err != nil {
// 					log.Fatalf("Unmarshal: %v", err)
// 				}
// 				fmt.Println("  ", providerRegions)
// 			} else {
// 				// sizes
// 				err = yaml.Unmarshal(yamlFile, providerSizes)
// 				if err != nil {
// 					log.Fatalf("Unmarshal: %v", err)
// 				}
// 				fmt.Println("  ", providerSizes)
// 			}

// 		}
// 	}
// try to use
// }

// main
func main() {

	// normally just process one provider's files at a time
	processDigitalOcean := flag.Bool("digitalocean", false, "process digitalocean files")
	processLinode := flag.Bool("linode", false, "process linode files")
	processVultr := flag.Bool("vultr", false, "process vultr files")
	findUbuntuImages := flag.Bool("listUbuntuImages", false, "list ubuntu images (no file processing)")

	flag.Parse()

	// make sure at least one is set
	if !*processDigitalOcean && !*processLinode && !*processVultr {
		fmt.Println("no provider specified.  see `process_files -h` for more details")
		os.Exit(1)
	}

	// create the importer we want to use
	var importer ProviderImporter
	if *processDigitalOcean {

		fmt.Println("processing DigitalOcean files")
		importer = new(DigitalOceanImporter)

	} else if *processLinode {

		fmt.Println("processing Linode files")
		importer = new(LinodeImporter)

	} else if *processVultr {
		fmt.Println("processing Vultr files")
		importer = new(VultrImporter)
	}

	if *findUbuntuImages {
		// don't really need to go through images as mappings already has
		// which one to use when creating a VM
		importer.FindUbuntuImages()

	} else {
		// only go through images and show which are ubuntu
		// this is a special mode so can set correct image in mappings
		success := importer.ReadMappings()
		if !success {
			os.Exit(1)
		}

		// go through provider VM sizes and filter for ones that
		// we will support
		importer.ConvertProviderImageSizes()

		// go through provider regions and filter for ones that
		// we will support
		importer.ConvertProviderRegions()

	}

}
