package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/eezhee/eezhee/cmd"
	homeDir "github.com/mitchellh/go-homedir"
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

// use ssh-keygen to get the fingerprint for a ssh key
func getSSHFingerprint() (string, error) {

	// TODO: check OS as this only works on linux & mac

	// run command and grab output
	// SSH_KEYGEN=`ssh-keygen -l -E md5 -f $HOME/.ssh/id_rsa`
	command := "ssh-keygen"
	dir, _ := homeDir.Dir()
	keyFile := dir + "/.ssh/id_rsa"
	cmd := exec.Command(command, "-l", "-E", "md5", "-f", keyFile)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	sshKeygenOutput := string(stdoutStderr)

	// take output and extract part we need
	// 2048 MD5:dc:8e:47:1f:42:cd:93:cf:8a:e2:19:4f:a1:02:3e:cf person@company.com (RSA)

	fields := strings.Split(sshKeygenOutput, " ")
	fingerprint := fields[1]

	// trim off the 'MD5:'
	fingerprint = fingerprint[4:]

	return fingerprint, nil
}

// check if ssh key already uploaded to DigitalOcean
func checkSSHKeyUploaded(fingerprint string) bool {

	// TODO: need to check if has been uploaded to DO
	// DO_SSH_KEYS=`doctl compute ssh-key list`
	// if [[ ${DO_SSH_KEYS} != *${DO_FINGERPRINT}* ]]; then
	// 	echo "Need to upload SSH key to DO"
	// 	# new_key_name should be same as $HOME
	// 	# doctl import new_key_name --public-key-file ~/.ssh/id_rsa.pub
	// fi

	return true
}

func buildVM() bool {

	// is there a deploy state file
	// yes, error out
	//

	sshFingerprint, err := getSSHFingerprint()
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(sshFingerprint)

	uploaded := checkSSHKeyUploaded(sshFingerprint)
	if !uploaded {
		return false
	}

	// get directory name
	// vmName := "eezhee-test"
	// imageName := "ubuntu-20-04-x64"
	// vmSize := "s-1vcpu-1gb"
	// region := "tor1"

	// // doctl compute droplet create
	// RESULT = `doctl compute droplet create $VM_NAME --image $VM_IMAGE --size $VM_SIZE --region $VM_REGION --ssh-keys $DO_FINGERPRINT -o json`

	return true
}

func teardown() bool {
	// read deploy.yaml
	// if no file, error out

	// get IP
	// try and find that VM with doctl tool
	// if no VM, error out

	// use doctl to delete the VM
	// remove deploy.yaml

	return true
}

func main() {

	cmd.Execute()

	// what command:
	//build
	buildVM()

	//list

	//teardown
	teardown()

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

	// TEARDOWN

	fmt.Print("done\n")
}
