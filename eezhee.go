package main

import (
	"github.com/eezhee/eezhee/cmd"
)

func main() {

	cmd.Execute()

	// provider digitalocean set apikey xxxx
	//                              default_size micro
	//                              default_region
	// provider digitalocean list regions
	// provider digitalocean list sizes
	// will add credentials to .credentials file
	// SEEMS LONGER.  Would be easier to justde `eezhee login digitalocean xxxxx`

	// VALIDATE
	// read file
	// see what's missing

	// EDIT
	// read file
	// display (prety print) hight what's missing
	// prompt for anything that's missing
	// allow to change / add items - possible?

	// DEPLOY
	// save deployment state (deploy.state)
	// use doctl to build a VM
	// use aws to build a VM
	// use gcloud to build a VM

	// use arkade (or helm) to install ingress & letsencrypt

	// defaults
	// region
	// vm size

}
