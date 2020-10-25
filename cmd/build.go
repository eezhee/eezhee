package cmd

import (
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
		err := buildCluster()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// buildVM will create a cluster
func buildCluster() error {

	// get app settings
	appConfig := config.NewAppConfig()
	err := appConfig.Load()
	if err != nil {
		return err
	}

	// make sure the cluster doesn't already exist
	// is there a deploy state file
	deployState := config.NewDeployState()
	if deployState.FileExists() {
		return errors.New("cluster already running (as per deploy-state file)")
	}

	// nope, so we are clear to create a new cluster

	// is there a deploy config file
	deployConfig := config.NewDeployConfig()
	if deployConfig.FileExists() {
		err := deployConfig.Load()
		if err != nil {
			// there is a file but we couldn't load it
			return err
		}
	}

	// set name for cluster - default to project & branch name
	if len(deployConfig.Name) == 0 {
		deployConfig.Name, _ = buildClusterName()
	}

	// get ssh key we will use to login to new VM
	sshFingerprint, err := getSSHFingerprint()
	if err != nil {
		return err
	}
	deployConfig.SSHFingerprint = sshFingerprint

	// see which cloud to create the VM on
	if len(deployConfig.Cloud) == 0 {
		deployConfig.Cloud = "digitalocean"
	}
	switch deployConfig.Cloud {
	case "digitalocean":
	// case "aws":
	// case "gcloud":
	// case "azure":
	default:
		return errors.New("only can deploy to digitalocean right now")
	}
	fmt.Println("deplying to", deployConfig.Cloud)

	// check what release of k3s should be used
	// default to stable if not set
	if len(deployConfig.K3sVersion) == 0 {
		deployConfig.K3sVersion = "stable"
	}

	// make sure it's valid and if its a pinned release or a channel,
	// convert to actual release to be installed
	// needs to be 'stable', 'latest', validChannelName (v1.19) or validReleaseName (v1.19.3)
	k3sManager := k3s.NewManager()
	release, err := k3sManager.Releases.Translate(deployConfig.K3sVersion)
	if err != nil {
		return err
	}
	deployConfig.K3sVersion = release

	// make sure we can talk to DigitalOcean
	DOManager := digitalocean.NewManager(appConfig.DigitalOceanAPIKey)

	// make sure this ssh key is loaded into DigitalOcean
	uploaded, err := DOManager.IsSSHKeyUploaded(sshFingerprint)
	if !uploaded {
		return err
	}

	// set rest of details for new VM
	if len(deployConfig.Region) == 0 {
		fmt.Println("selecting closest region")
		deployConfig.Region, err = DOManager.SelectClosestRegion()
		if err != nil {
			return err
		}
		fmt.Println(deployConfig.Region, "is closest")
	}

	if len(deployConfig.Size) == 0 {
		deployConfig.Size = "s-1vcpu-1gb"
	}
	imageName := "ubuntu-20-04-x64"

	// time to create the VM
	fmt.Println("creating a VM")
	vmInfo, err := DOManager.CreateVM(
		deployConfig.Name, imageName, deployConfig.Size,
		deployConfig.Region, deployConfig.SSHFingerprint,
	)
	if err != nil {
		return err
	}
	vmID := vmInfo.ID
	status := vmInfo.Status

	// see if vm ready.  if not need to wait as don't have IP yet
	for strings.Compare(status, "active") != 0 {

		// wait a bit
		time.Sleep(2 * time.Second)
		vmInfo, err = DOManager.GetVMInfo(vmID)
		if err != nil {
			// TODO: vm has been created, really should delete (or should we add retry to getVMInfo?)
			return err
		}
		status = vmInfo.Status
	}
	vmPublicIP, err := vmInfo.GetPublicIP()
	if err != nil {
		return err
	}

	// pause as ssh might not be ready
	time.Sleep(2 * time.Second)

	fmt.Println("VM is ready")

	// save current state
	deployState.Cloud = deployConfig.Cloud
	deployState.ID = vmInfo.ID
	deployState.Name = vmInfo.Name
	deployState.Region = vmInfo.Region.Slug
	deployState.Size = vmInfo.Size.Slug
	deployState.IP = vmPublicIP
	deployState.SSHFingerprint = deployConfig.SSHFingerprint
	err = deployState.Save()
	if err != nil {
		return err
	}

	// install k3s on the VM
	k3sVersion := deployConfig.K3sVersion
	fmt.Println("installing k3s release", k3sVersion)
	k3sManager.Install(vmPublicIP, k3sVersion, deployConfig.Name)

	// done, cluster up and running

	// update state file & re-save
	deployState.K3sVersion = k3sVersion
	err = deployState.Save()
	if err != nil {
		return err
	}

	return nil
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
