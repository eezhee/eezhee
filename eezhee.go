package main

import (
	"github.com/eezhee/eezhee/cmd"
)

func generateAppName() string {
	return "appname"
}

func getCurrentGitBranch() string {

	// check for .git subdir
	// this only works in root dir of project (so skip for now)

	// GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`

	//   if [ $GIT_BRANCH == 'master' ]; then
	//     BRANCH=''
	//   else
	//     BRANCH=${GIT_BRANCH}-
	//   fi
	// else
	//   BRANCH=''
	// fi
	// VM_NAME=${APP_NAME}-${BRANCH}cluster

	return "master"
}

func main() {

	cmd.Execute()

	// is there a deploy
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

}
