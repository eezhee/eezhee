package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type LinodeImporter struct {
}

// findUbuntuImages - go through images and find ubuntu base images
func (do *LinodeImporter) FindUbuntuImages() bool {

	// read in the yaml
	filename := "./raw/" + "linode-images.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	images := result["images"].([]interface{})

	fmt.Printf("  images file has %d images\n", len(images))
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
				fmt.Printf("    ID: %-9s  Slug: %-20s  Name: %-20s   Created: %s\n", imageID, slug, imageName, createdAt)

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
func (do *LinodeImporter) ConvertProviderImageSizes() bool {

	// read in the yaml
	filename := "./raw/" + "linode-sizes.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonFile), &result)
	sizes := result["sizes"].([]interface{})

	fmt.Printf("  sizes file has %d sizes\n", len(sizes))

	for _, size := range sizes {

		sizeInfo := size.(map[string]interface{})
		slug := sizeInfo["slug"].(string)
		available := sizeInfo["available"].(bool)
		if available {

			processors := int(sizeInfo["vcpus"].(float64))
			memory := int(sizeInfo["memory"].(float64) / 1024)
			disk := int(sizeInfo["disk"].(float64))
			// transfer := int(sizeInfo["transfer"].(float64))
			// description := sizeInfo["description"].(string)
			// regions
			// price_hourly
			// price_monthly

			// get a list of all fields
			// for key, value := range sizeInfo {
			// 	fmt.Println(key, ": ", value)
			// }

			// convert to eezhee format
			fmt.Printf("    %s: (cpu: %d mem: %d disk: %d)\n", slug, processors, memory, disk)

		} else {
			fmt.Printf("    %s is not available\n", slug)
		}

	}

	// save eezhee formated data to file

	return true
}

// convertProviderImageSizes will convert a provider json file to eezhee format
func (do *LinodeImporter) ConvertProviderRegions() bool {

	// read in the yaml
	filename := "./raw/" + "linode-regions.json"
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
		available := regionInfo["available"].(bool)
		if available {
			slug := regionInfo["slug"].(string)
			name := regionInfo["name"].(string)

			// sizes
			// features
			fmt.Printf("    %s (%s)\n", slug, name)

			// get a list of all fields
			// for key, value := range regionInfo {
			// 	fmt.Println(key, ": ", value)
			// }

		}

		// convert to eezhee format
	}

	// save eezhee formated data to file

	return true
}
