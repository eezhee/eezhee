package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type DigitalOceanImporter struct {
	Mappings ProviderMappings
}

// findUbuntuImages - go through images and find ubuntu base images
// this can be used to find the image that Eezhee should use to build VMs
func (do *DigitalOceanImporter) FindUbuntuImages() bool {

	// read in the yaml
	filename := DATA_PATH + "digitalocean-images.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonFile), &result)
	if err != nil {
		log.Printf("could not parse %s: #%v ", filename, err)
		return false
	}
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

				// get a list of all fields - for debugging only
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
func (do *DigitalOceanImporter) ConvertProviderImageSizes() bool {

	// read in the raw data
	filename := DATA_PATH + "digitalocean-sizes.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonFile), &result)
	if err != nil {
		log.Printf("could not parse %s: #%v ", filename, err)
		return false
	}
	sizes := result["sizes"].([]interface{})

	fmt.Printf("  sizes file has %d sizes\n", len(sizes))

	// go through each VM size and filter out just the ones that Eezhee
	// supports.  Use the Mappings structure to do so
	for _, size := range sizes {

		// get basic info on this VM size
		sizeInfo := size.(map[string]interface{})

		// can new VMs be made with this size?
		available := sizeInfo["available"].(bool)
		if !available {
			// fmt.Printf("    %s is not available\n", slug)
			continue
		}

		// is this size one we want to support
		slug := sizeInfo["slug"].(string)
		_, sizeToBeMapped := do.Mappings.Sizes[slug]
		if sizeToBeMapped {

			processors := int(sizeInfo["vcpus"].(float64))
			memory := int(sizeInfo["memory"].(float64) / 1024)
			disk := int(sizeInfo["disk"].(float64))
			transfer := int(sizeInfo["transfer"].(float64))
			description := sizeInfo["description"].(string)
			price_hourly := sizeInfo["price_hourly"].(float64)
			price_monthly := sizeInfo["price_monthly"].(float64)
			// regions

			// convert to eezhee format
			fmt.Printf("    %s: (cpu: %d mem: %d disk: %d xfer: %d)\n", slug, processors, memory, disk, transfer)
			fmt.Printf("        (descr: %s $/hr: %v $/mth: %v)\n", description, price_hourly, price_monthly)
		}
	}

	// save eezhee formated data to file

	return true
}

// convertProviderImageSizes will convert a provider json file to eezhee format
func (do *DigitalOceanImporter) ConvertProviderRegions() bool {

	// read in the raw data
	filename := DATA_PATH + "digitalocean-regions.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonFile), &result)
	if err != nil {
		log.Printf("could not parse %s: #%v ", filename, err)
		return false
	}
	regions := result["regions"].([]interface{})

	fmt.Printf("  regions file has %d regoins\n", len(regions))

	// go through each region and filter out just the ones that Eezhee
	// supports.  Use the Mappings structure to do so
	for _, region := range regions {

		regionInfo := region.(map[string]interface{})

		// only look at regions provider says are available
		available := regionInfo["available"].(bool)
		if available {

			// check if we should use this region
			slug := regionInfo["slug"].(string)
			_, regionToBeMapped := do.Mappings.Regions[slug]
			if regionToBeMapped {

				name := regionInfo["name"].(string)
				// sizes
				// features
				fmt.Printf("    %s (%s)\n", slug, name)

				// convert to eezhee format
			}
		}
	}

	// save eezhee formated data to file

	return true
}

func (do *DigitalOceanImporter) ReadMappings() bool {

	// read in the data
	filename := "./digitalocean-mappings.json"
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	err = json.Unmarshal(jsonFile, &do.Mappings)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// use this to process the provider data
	fmt.Println(do.Mappings.Image)

	return true
}
