// sample code to read in provider data

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var providers = []string{"digitalocean", "linode", "vultr"}
var fileTemplates = []string{"-region-mappings.yaml", "-sizes-mapping.yaml"}

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

// go through images and find ubuntu base images
func findUbuntuImages() bool {

	// read in the yaml
	filename := "./raw/" + "digitalocean" + "-images.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	images := result["images"].([]interface{})
	for _, image := range images {

		imageInfo := image.(map[string]interface{})
		status := imageInfo["status"].(string)
		public := imageInfo["public"].(bool)

		// only care about images that are available
		if (status == "available") && public {

			// skip bad slug names
			if imageInfo["slug"] == nil {
				// shouldn't really happen but DO files sometimes have slug set to 'null'
				continue
			}

			// only want Ubuntu based distributions
			slug := imageInfo["slug"].(string)
			if strings.HasPrefix(slug, "ubuntu") {

				imageName := imageInfo["name"].(string)
				imageID := strconv.FormatFloat(imageInfo["id"].(float64), 'f', 0, 64)
				// description := imageInfo["description"].(string)
				createdAt := imageInfo["created_at"].(string)
				// distribution := imageInfo["distribution"].(string)
				fmt.Printf("ID: %-9s  Slug: %-20s  Name: %-20s   Created: %s\n", imageID, slug, imageName, createdAt)

				// get a list of all fields
				// for key, value := range imageInfo {
				// 	// regions
				// 	fmt.Println(key, ": ", value)
				// }

			}

		}
	}

	return true
}

// convertProviderImageSizes will convert a provider json file to eezhee format
func convertProviderImageSizes() bool {

	fmt.Println("=======")

	// read in the yaml
	filename := "./raw/" + "digitalocean" + "-sizes.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	sizes := result["sizes"].([]interface{})
	for _, size := range sizes {

		sizeInfo := size.(map[string]interface{})

		// get a list of all fields
		for key, value := range sizeInfo {
			// regions
			fmt.Println(key, ": ", value)
		}

		// convert to eezhee format
		fmt.Println("----")
	}

	// save eezhee formated data to file

	return true
}

// convertProviderImageSizes will convert a provider json file to eezhee format
func convertProviderRegions() bool {

	fmt.Println("=======")

	// read in the yaml
	filename := "./raw/" + "digitalocean" + "-regions.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	regions := result["regions"].([]interface{})
	for _, region := range regions {

		regionInfo := region.(map[string]interface{})

		// get a list of all fields
		for key, value := range regionInfo {
			fmt.Println(key, ": ", value)
		}

		// convert to eezhee format
		fmt.Println("----")
	}

	// save eezhee formated data to file

	return true
}

// stories
// 1 given a user's location, return the closest region
// IPToLocation returns country and city/state?
// ListCloset()
// 2 given a user's preferred country (ie US), return the best region
// eezhee regions list - will list regions for each provider (both id and cloud_id)

func main() {

	// we really don't need to parse images as we have
	// image slug hard coded in eezhee but will be handy
	// if want to update to newer version of ubuntu
	findUbuntuImages()
	// take raw json files from provider and convert to eezhee common format

	convertProviderImageSizes()
	convertProviderRegions()

	os.Exit(1)

	// THIS SHOULD GO IN EEZHEE
	// read in yaml file
	for _, provider := range providers {
		for _, template := range fileTemplates {

			filename := "./" + provider + template
			fmt.Println("filename:", filename)

			// read in the yaml
			yamlFile, err := ioutil.ReadFile(filename)
			if err != nil {
				log.Printf("yamlFile.Get err   #%v ", err)
				continue
			}

			var providerRegions Regions
			var providerSizes InstanceSizes

			if strings.Contains(filename, "region") {
				// regions
				err = yaml.Unmarshal(yamlFile, providerRegions)
				if err != nil {
					log.Fatalf("Unmarshal: %v", err)
				}
				fmt.Println("  ", providerRegions)
			} else {
				// sizes
				err = yaml.Unmarshal(yamlFile, providerSizes)
				if err != nil {
					log.Fatalf("Unmarshal: %v", err)
				}
				fmt.Println("  ", providerSizes)
			}

		}
	}
	// try to use

}
