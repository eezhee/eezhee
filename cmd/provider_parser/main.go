package main

import (
	"os"

	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/eezhee/eezhee/pkg/linode"
	"github.com/eezhee/eezhee/pkg/vultr"
	log "github.com/sirupsen/logrus"
)

func main() {

	// load the list of providers

	// load digitalocean data
	// transform the data
	_, err := digitalocean.NewManager("api_key")
	// DOClient, err := digitalocean.NewManager("api_key")
	if err != nil {
		log.Error("could not load digitalocean api client")
		os.Exit(1)
	}
	// get regions

	// LinodeClient := linode.NewManager("api_key")
	_, err = linode.NewManager("api_key")
	if err != nil {
		log.Error("could not load linode api client")
		os.Exit(2)
	}
	// get regions

	// VultrClient, err := vultr.NewManager("api_key")
	_, err = vultr.NewManager("api_key")
	if err != nil {
		log.Error("could not load vultr api client")
		os.Exit(3)
	}
	// get regions

}
