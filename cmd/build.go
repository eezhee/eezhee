package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/eezhee/eezhee/pkg/config"
	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/eezhee/eezhee/pkg/k3s"

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

// TODO move to digitalocean struct!!
func getVMPublicIP(vmInfo digitalocean.VMInfo) (publicIP string, err error) {

	// go through all network and find which one is public
	numNetworks := len(vmInfo.Networks.V4Info)

	for i := 0; i < numNetworks; i++ {

		networkType := vmInfo.Networks.V4Info[i].Type

		if strings.Compare(networkType, "public") == 0 {
			publicIP := vmInfo.Networks.V4Info[i].IPAddress
			return publicIP, nil
		}

	}

	// did not find public IP
	return publicIP, errors.New("VM does not have public IP")
}

// buildVM will create a cluster
func buildVM() (bool, error) {

	// make sure the cluster doesn't already exist
	// is there a deploy state file
	deployState := config.NewDeployState()
	if deployState.FileExists() {
		return false, errors.New("cluster already running (as per deploy-state file)")
	}

	// nope, so we are clear to create a new cluster

	// is there a deploy config file
	deployConfig := config.NewDeployConfig()
	if deployConfig.FileExists() {
		err := deployConfig.Load()
		if err != nil {
			// there is a file but we couldn't load it
			fmt.Println(err)
			return false, err
		}
	}

	// set name for cluster - default to project & branch name
	if len(deployConfig.Name) == 0 {
		deployConfig.Name, _ = buildClusterName()
	}

	// get ssh key we will use to login to new VM
	sshFingerprint, err := getSSHFingerprint()
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	// figure out which version of k3s to install
	k3sManager := k3s.NewManager()
	k3sVersion := k3sManager.GetLatestVersion()
	fmt.Println(k3sVersion)

	// make sure we can talk to DigitalOcean
	DOManager := digitalocean.NewManager()
	haveRequirements, err := DOManager.CheckRequirements()
	if !haveRequirements {
		fmt.Println(err)
		return false, err
	}

	// make sure this ssh key is loaded into DigitalOcean
	uploaded, err := DOManager.CheckSSHKeyUploaded(sshFingerprint)
	if !uploaded {
		fmt.Println(err)
		return false, err
	}

	// set rest of details for new VM
	if len(deployConfig.Region) == 0 {
		deployConfig.Region, err = DOManager.SelectClosestRegion()
		if err != nil {
			return false, err
		}
	}
	if len(deployConfig.Size) == 0 {
		deployConfig.Size = "s-1vcpu-1gb"
	}
	imageName := "ubuntu-20-04-x64"

	// vm := DOManager.CreateVM()

	// time to create the VM
	var vmInfo []digitalocean.VMInfo

	//TODO: add tags '--tag-names' such as 'eezhee' 'userName'
	cmd := exec.Command("doctl", "compute", "droplet", "create", deployConfig.Name,
		"--image", imageName, "--size", deployConfig.Size, "--region", deployConfig.Region, "--ssh-keys", sshFingerprint,
		"--tag-name", "eezhee", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(stdoutStderr))
		fmt.Println(err)
		return false, err
	}
	fmt.Println("vm is being built")

	// parse the json output
	json.Unmarshal([]byte(stdoutStderr), &vmInfo)

	vmID := vmInfo[0].ID
	status := vmInfo[0].Status

	// see if vm ready.  if not need to wait as don't have IP yet
	for strings.Compare(status, "active") != 0 {

		// wait a bit
		time.Sleep(2 * time.Second)
		vmInfo, err = DOManager.GetVMInfo(vmID)
		if err != nil {
			fmt.Println(err)
			// TODO: vm has been created, really should delete (or should we add retry to getVMInfo?)
			return false, err
		}
		status = vmInfo[0].Status
	}
	fmt.Println("vm now ready to use")

	// save key details in state file
	deployState.Cloud = "digitalocean"
	deployState.ID = vmInfo[0].ID
	deployState.Name = vmInfo[0].Name
	deployState.Region = vmInfo[0].Region.Slug
	deployState.Size = vmInfo[0].SizeSlug
	publicIP, err := getVMPublicIP(vmInfo[0])
	// if err != nil {
	// should never happen - if here, but in DO API
	// }
	deployState.IP = publicIP
	deployState.Save()

	// time to install k3s on the new VM
	k3sManager.Install()
	// k3sManager.Install(latestRelease)

	// add k3s tag to VM

	return true, nil
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

	// check for invalid characters
	clusterName = strings.ReplaceAll(clusterName, "_", "-")

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
