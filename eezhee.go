package main

import (
	"github.com/eezhee/eezhee/cmd"
	"github.com/eezhee/eezhee/pkg/k3d"
)

func main() {

	k3d.CreateK3dCluster()

	cmd.Execute()

	// LOGIN
	// will add credentials to .credentials file

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
