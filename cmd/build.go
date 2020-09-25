package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Print the version number of Eezhee",
	Long:  `All software has versions. This is Eezhee's`,
	Run: func(cmd *cobra.Command, args []string) {
		buildVM()
	},
}

// check if doctl is install and we have an auth token
func validateRequirements() (bool, error) {

	// is doctl installed
	cmd := exec.Command("which", "doctl")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.New("doctl is not installed")
	}

	// has user authenticated
	// should list at least one item, normally is `default` context
	cmd = exec.Command("doctl", "auth", "list")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return false, errors.New("doctl is not logged in.  Use 'doctl auth init'")
	}
	// fmt.Println(string(stdoutStderr))

	return true, nil
}

func buildVM() bool {

	// is there a deploy state file

	haveRequirements, err := validateRequirements()
	if !haveRequirements {
		fmt.Println(err)
		return false
	}

	vmName, _ := buildClusterName()
	fmt.Println(vmName)

	sshFingerprint, err := getSSHFingerprint()
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(sshFingerprint)

	uploaded, err := checkSSHKeyUploaded(sshFingerprint)
	if !uploaded {
		fmt.Println(err)
		return false
	}

	// get directory name
	imageName := "ubuntu-20-04-x64"
	vmSize := "s-1vcpu-1gb"
	region := "tor1"

	// time to create the VM
	cmd := exec.Command("doctl", "compute", "droplet", "create", vmName, "--image", imageName, "--size", vmSize, "--region", region, "--ssh-keys", sshFingerprint, "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(stdoutStderr)

	// parse the output
	// get the ID
	// save the ID

	return true
}

// figure out what to call k3s cluster
// based on combo of app name and git branch (if not master
// eg webapp, webapp-staging, webapp-newFeatureBranch)
func buildClusterName() (string, error) {

	clusterName := ""

	appName, _ := getCurrentDir()
	branchName, err := getGitBranchName()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if len(branchName) > 0 {
		clusterName = appName + "-" + branchName
	} else {
		clusterName = appName + branchName
	}
	// fmt.Println(clusterName)

	return clusterName, nil
}

func getCurrentDir() (string, error) {

	fullDirPath, _ := os.Getwd()
	fields := strings.Split(fullDirPath, "/")
	currentDir := fields[len(fields)-1]

	return currentDir, nil
}

func getGitBranchName() (string, error) {

	// see if there is a .git subdir
	command := "git"
	cmd := exec.Command(command, "rev-parse", "--abbrev-ref", "HEAD")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	branchName := string(stdoutStderr)
	branchName = strings.TrimSpace(branchName)

	if strings.Compare(branchName, "master") == 0 {
		return "", nil
	}

	return branchName, nil
}

// use ssh-keygen to get the fingerprint for a ssh key
func getSSHFingerprint() (string, error) {

	// TODO: check OS as this only works on linux & mac

	// run command and grab output
	// SSH_KEYGEN=`ssh-keygen -l -E md5 -f $HOME/.ssh/id_rsa`
	command := "ssh-keygen"
	dir, _ := homedir.Dir()
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

type digitalOceanSSHKeys struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
}

// check if ssh key already uploaded to DigitalOcean
func checkSSHKeyUploaded(fingerprint string) (bool, error) {

	var sshKeys []digitalOceanSSHKeys

	// get list of sshkeys DO knows about
	cmd := exec.Command("doctl", "compute", "ssh-key", "list", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	json.Unmarshal([]byte(stdoutStderr), &sshKeys)

	// go through each key and see if it matches what is on this machine
	for i := 0; i < len(sshKeys); i++ {
		if strings.Compare(fingerprint, sshKeys[i].Fingerprint) == 0 {
			// fmt.Println("found ssh key")
			return true, nil
		}
	}

	return false, errors.New("ssh key not available.  you need to upload it to DO using the web console")
}
