package main

import "fmt"

func main() {

	// read $HOME/.eezhee/.credentials

	// read deployment file (deploy.yaml)
	// cloud to use (if not specified use whatever credentials we have)
	// host name & dns provider

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

	// use k3sup to install k3s

	// use arkade (or helm) to install ingress & letsencrypt

	// update DNS

	// defaults
	// region
	// vm size

	// TEARDOWN

	fmt.Print("test")
}
