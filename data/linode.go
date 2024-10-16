package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type LinodeImporter struct {
	Mappings ProviderMappings
}

// findUbuntuImages - go through images and find ubuntu base images
func (l *LinodeImporter) FindUbuntuImages() bool {

	// read in the yaml
	filename := DATA_PATH + "linode-images.json"
	jsonFile, err := os.ReadFile(filename)
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
	images := result["data"].([]interface{})

	fmt.Printf("  images file has %d images\n", len(images))
	for _, image := range images {

		// "deprecated" : false,
		// "is_public" : true,
		// "vendor" : "Alpine"
		// "id" : "linode/alpine3.10",
		// "label" : "Alpine 3.10",

		// "created" : "2019-06-20T17:17:11",
		// "created_by" : "linode",
		// "description" : "",
		// "eol" : "2021-05-01T04:00:00",
		// "expiry" : null,
		// "size" : 300,
		// "type" : "manual",

		imageInfo := image.(map[string]interface{})

		deprecated := imageInfo["deprecated"].(bool)
		is_public := imageInfo["is_public"].(bool)

		// only care about images that are available
		if is_public && !deprecated {

			// only want Ubuntu based distributions
			vendor := imageInfo["vendor"].(string)
			if strings.Compare(vendor, "Ubuntu") == 0 {

				id := imageInfo["id"].(string)
				label := imageInfo["label"].(string)

				// description := imageInfo["description"].(string)
				// createdAt := imageInfo["created_at"].(string)
				// distribution := imageInfo["distribution"].(string)
				fmt.Printf("    ID: %-9s  Label: %-20s\n", id, label)

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
func (l *LinodeImporter) ConvertProviderImageSizes() bool {

	// read in the yaml
	filename := DATA_PATH + "linode-types.json"
	jsonFile, err := os.ReadFile(filename)
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
	sizes := result["data"].([]interface{})

	fmt.Printf("  sizes file has %d sizes\n", len(sizes))

	for _, size := range sizes {

		sizeInfo := size.(map[string]interface{})

		id := sizeInfo["id"].(string)
		_, sizeToBeMapped := l.Mappings.Sizes[id]
		if sizeToBeMapped {
			label := sizeInfo["label"].(string)

			processors := int(sizeInfo["vcpus"].(float64))
			memory := int(sizeInfo["memory"].(float64) / 1024)
			disk := int(sizeInfo["disk"].(float64) / 1024)
			// transfer := int(sizeInfo["transfer"].(float64) / 1000)

			//  "network_out" : 1000,
			// 	"addons" : {
			// 		"backups" : {
			// 			 "price" : {
			// 					"hourly" : 0.003,
			// 					"monthly" : 2
			// 			 }
			// 		}
			//  },
			//  "class" : "nanode",
			//  "gpus" : 0,
			//  "price" : {
			// 		"hourly" : 0.0075,
			// 		"monthly" : 5
			//  },
			//  "successor" : null,

			// get a list of all fields
			// for key, value := range sizeInfo {
			// 	fmt.Println(key, ": ", value)
			// }

			// convert to eezhee format
			fmt.Printf("    %s(%s): (cpu: %d mem: %d disk: %d)\n", id, label, processors, memory, disk)

		}

	}

	// save eezhee formated data to file

	return true
}

// convertProviderImageSizes will convert a provider json file to eezhee format
func (l *LinodeImporter) ConvertProviderRegions() bool {

	// read in the yaml
	filename := DATA_PATH + "linode-regions.json"
	jsonFile, err := os.ReadFile(filename)
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
	regions := result["data"].([]interface{})

	fmt.Printf("  regions file has %d regoins\n", len(regions))

	for _, region := range regions {

		regionInfo := region.(map[string]interface{})

		status := regionInfo["status"].(string)
		if status == "ok" {

			// check if we should use this region
			id := regionInfo["id"].(string)
			_, regionToBeMapped := l.Mappings.Regions[id]
			if regionToBeMapped {
				country := regionInfo["country"].(string)
				// 	"capabilities" : [
				// 		"Linodes",
				// 		"NodeBalancers",
				// 		"Block Storage",
				// 		"GPU Linodes",
				// 		"Kubernetes",
				// 		"Cloud Firewall"
				//  ],
				//  "resolvers" : {
				// 		"ipv4" : "172.105.34.5,172.105.35.5,172.105.36.5,172.105.37.5,172.105.38.5,172.105.39.5,172.105.40.5,172.105.41.5,172.105.42.5,172.105.43.5",
				// 		"ipv6" : "2400:8904::f03c:91ff:fea5:659,2400:8904::f03c:91ff:fea5:9282,2400:8904::f03c:91ff:fea5:b9b3,2400:8904::f03c:91ff:fea5:925a,2400:8904::f03c:91ff:fea5:22cb,2400:8904::f03c:91ff:fea5:227a,2400:8904::f03c:91ff:fea5:924c,2400:8904::f03c:91ff:fea5:f7e2,2400:8904::f03c:91ff:fea5:2205,2400:8904::f03c:91ff:fea5:9207"
				//  },

				fmt.Printf("    %s (%s)\n", id, country)
			}
		}

		// convert to eezhee format
	}

	// save eezhee formated data to file

	return true
}

func (l *LinodeImporter) ReadMappings() bool {
	// read in the data
	filename := "./linode-mappings.json"
	jsonFile, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("jsonFile.Readfile error: #%v ", err)
		return false
	}

	// parse the file
	err = json.Unmarshal(jsonFile, &l.Mappings)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// use this to process the provider data
	// fmt.Println(l.Mappings.Image)
	// fmt.Println(l.Mappings.Sizes)

	return true
}
