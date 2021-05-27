package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type VultrImporter struct {
	Mappings ProviderMappings
}

// findUbuntuImages - go through images and find ubuntu base images
func (v *VultrImporter) FindUbuntuImages() bool {

	// read in the yaml
	filename := DATA_PATH + "vultr-os.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	images := result["os"].([]interface{})

	fmt.Printf("  images file has %d images\n", len(images))
	for _, image := range images {

		imageInfo := image.(map[string]interface{})

		// "id" : 124,
		// "name" : "Windows 2012 R2 x64"
		// "family" : "windows",
		// "arch" : "x64",

		family := imageInfo["family"].(string)

		// only want Ubuntu based distributions
		if family == "ubuntu" {

			id := int(imageInfo["id"].(float64))
			arch := imageInfo["arch"].(string)
			name := imageInfo["name"].(string)

			fmt.Printf("    ID: %d  Name: %-20s  arch: %-10s\n", id, name, arch)

			// get a list of all fields
			// for key, value := range imageInfo {
			// 	// regions
			// 	fmt.Println(key, ": ", value)
			// }

		}
	}

	return true
}

// convertProviderImageSizes will convert a provider json file to eezhee format
func (v *VultrImporter) ConvertProviderImageSizes() bool {

	// read in the yaml
	filename := DATA_PATH + "vultr-plans.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	sizes := result["plans"].([]interface{})

	fmt.Printf("  sizes file has %d sizes\n", len(sizes))

	for _, size := range sizes {

		sizeInfo := size.(map[string]interface{})

		id := sizeInfo["id"].(string)

		processors := int(sizeInfo["vcpu_count"].(float64))
		memory := int(sizeInfo["ram"].(float64) / 1024)
		disk := int(sizeInfo["disk"].(float64))
		// transfer := int(sizeInfo["bandwidth"].(float64))
		// "disk_count" : 1,
		// "locations" : [
		// 	 "sgp"
		// ],
		// "monthly_cost" : 5,
		// "type" : "vc2",

		// get a list of all fields
		// for key, value := range sizeInfo {
		// 	fmt.Println(key, ": ", value)
		// }

		// convert to eezhee format
		fmt.Printf("    %s: (cpu: %d mem: %d disk: %d)\n", id, processors, memory, disk)

	}

	// save eezhee formated data to file

	return true
}

// convertProviderImageSizes will convert a provider json file to eezhee format
func (v *VultrImporter) ConvertProviderRegions() bool {

	// read in the yaml
	filename := DATA_PATH + "vultr-regions.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	regions := result["regions"].([]interface{})

	fmt.Printf("  regions file has %d regoins\n", len(regions))

	for _, region := range regions {

		regionInfo := region.(map[string]interface{})

		id := regionInfo["id"].(string)

		city := regionInfo["city"].(string)
		country := regionInfo["country"].(string)
		continent := regionInfo["continent"].(string)

		fmt.Printf("    %s (%s, %s, %s)\n", id, city, country, continent)

		// "options" : [
		// 	 "ddos_protection"
		// ]

		// get a list of all fields
		// for key, value := range regionInfo {
		// 	fmt.Println(key, ": ", value)
		// }

		// convert to eezhee format
	}

	// save eezhee formated data to file

	return true
}

func (v *VultrImporter) ReadMappings() bool {
	// read in the data
	filename := "./vultr-mappings.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	err = json.Unmarshal(jsonFile, &v.Mappings)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// use this to process the provider data
	fmt.Println(v.Mappings.Image)
	fmt.Println(v.Mappings.Sizes)

	return true
}
