package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/eezhee/eezhee/pkg/aws"
	"github.com/eezhee/eezhee/pkg/config"
	"github.com/eezhee/eezhee/pkg/core"
	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/eezhee/eezhee/pkg/k3s"
	"github.com/eezhee/eezhee/pkg/linode"
	"github.com/eezhee/eezhee/pkg/vultr"
	"golang.org/x/crypto/ssh"

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

	// see which cloud we have an api token for
	defaultCloud := appConfig.GetDefaultCloud()
	if len(defaultCloud) == 0 {
		// opps, no api keys specified so can't proceed until resolved
		return errors.New("no cloud provider configured. User 'eezhee auth add'")
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

	// make sure we have a name for the cluster
	// if not set, create a name
	if len(deployConfig.Name) == 0 {
		deployConfig.Name, _ = buildClusterName()
	}

	// load ssh key we will use
	var sshKey core.SSHKey

	dir, _ := homedir.Dir()
	keyFile := dir + "/.ssh/id_rsa.pub"

	err = sshKey.LoadPublicKey(keyFile)
	if err != nil {
		return err
	}
	deployConfig.SSHPublicKey = sshKey.GetPublicKey()

	// does config specify which cloud to use
	// if not, use one that we have credentials for
	if len(deployConfig.Cloud) == 0 {
		deployConfig.Cloud = defaultCloud
	}

	// make sure we have a valid cloud
	switch deployConfig.Cloud {
	case "digitalocean", "linode", "vultr":
	// case "gcloud":
	// case "azure":
	// case "aws":
	default:
		return errors.New("no or invalid cloud specified")
	}
	fmt.Println("deplying to", deployConfig.Cloud)

	// has a release of k3s been specified?
	// if not, use latest stable release
	if len(deployConfig.K3sVersion) == 0 {
		deployConfig.K3sVersion = "stable"
	}

	// we're pretty flexibly in how release is specified.
	// could be 'stable', 'latest', validChannelName (v1.19) or validReleaseName (v1.19.3)
	// as a result, we need to translate it to exactly which release to install
	k3sManager := k3s.NewManager()
	release, err := k3sManager.Releases.Translate(deployConfig.K3sVersion)
	if err != nil {
		return err
	}
	deployConfig.K3sVersion = release

	// ok validation completed, time to get building

	// create an instance of the VM manager which does building
	var vmManager core.VMManager

	switch deployConfig.Cloud {
	case "aws":
		// TODO: work out how to authenticate for aws
		vmManager = aws.NewManager(appConfig.LinodeAPIKey)
		if vmManager == nil {
			return errors.New("could not create aws client")
		}
	case "digitalocean":
		vmManager = digitalocean.NewManager(appConfig.DigitalOceanAPIKey)
		if vmManager == nil {
			return errors.New("could not create digitalocean client")
		}
	case "linode":
		vmManager = linode.NewManager(appConfig.LinodeAPIKey)
		if vmManager == nil {
			return errors.New("could not create linode client")
		}
	case "vultr":
		vmManager = vultr.NewManager(appConfig.VultrAPIKey)
		if vmManager == nil {
			return errors.New("could not create vultr client")
		}
	default:
		// should never get here (but lets play it safe)
		return errors.New("invalid cloud type")
	}

	// TODO: for DO, should upload it if not there yet
	// make sure this ssh key is loaded into cloud platform
	_, err = vmManager.IsSSHKeyUploaded(sshKey)
	if err != nil {
		return err
	}

	// TODO: need to valide if it a valid region for given cloud
	// set rest of details for new VM
	if len(deployConfig.Region) == 0 {
		fmt.Println("selecting closest region")
		deployConfig.Region, err = vmManager.SelectClosestRegion()
		if err != nil {
			return err
		}
		fmt.Println(deployConfig.Region, "is closest")
	}

	// TODO - allow config to specify size/type
	// TODO - translate generic size/type to provider specific
	var imageName string
	switch deployConfig.Cloud {
	case "digitalocean":
		if len(deployConfig.Size) == 0 {
			deployConfig.Size = "s-1vcpu-1gb"
		}
		imageName = "ubuntu-20-04-x64"
	case "linode":
		if len(deployConfig.Size) == 0 {
			deployConfig.Size = "g6-nanode-1"
		}
		imageName = "linode/ubuntu20.04"
	case "vultr":
		if len(deployConfig.Size) == 0 {
			deployConfig.Size = "201" // $5/month
		}
		imageName = "387" // ubuntu 20.04
	}

	// time to create the VM
	fmt.Println("creating a VM")
	vmInfo, err := vmManager.CreateVM(
		deployConfig.Name, imageName, deployConfig.Size,
		deployConfig.Region, sshKey,
	)
	if err != nil {
		return err
	}
	vmID := vmInfo.ID
	status := vmInfo.Status

	// see if vm ready.  if not need to wait as don't have IP yet

	// all providers have their own status messages
	// the only one we standardize is the final one
	// provider needs to convert to "running"

	lastStatus := ""
	for strings.Compare(status, "running") != 0 {

		// wait a bit
		time.Sleep(2 * time.Second)
		vmInfo, err = vmManager.GetVMInfo(vmID)
		if err != nil {
			// TODO: vm has been created, really should delete (or should we add retry to getVMInfo?)
			return err
		}
		status = vmInfo.Status

		// print status if it has changed since last time
		if strings.Compare(lastStatus, status) != 0 {
			fmt.Println(status)
			lastStatus = status
		}
	}

	// some VMs have multiple IPs (internal and public)
	// we just need the public one
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
	// TODO save public key
	deployState.SSHPublicKey = deployConfig.SSHPublicKey
	err = deployState.Save()
	if err != nil {
		return err
	}

	// TODO: refactor this into another function

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

	// TODO: what about installing ingress?

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

// LoadSSHPublicKey will load a ssh key
func LoadSSHPublicKey() {

	dir, _ := homedir.Dir()
	keyFile := dir + "/.ssh/id_rsa.pub"
	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return
	}
	publicKey := string(data)
	fmt.Println(publicKey)

	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(data))
	if err != nil {
		panic(err)
	}
	fmt.Println(pk)
	original := string(ssh.MarshalAuthorizedKey(pk))
	fmt.Println(original)

	if strings.Compare(publicKey, original) == 0 {
		// can't do this as publicKey has email address at end of string
		fmt.Println("can go back and forth")
	}
	fmt.Println(pk.Type())

	fingerprint := ssh.FingerprintLegacyMD5(pk)
	fmt.Println(fingerprint)

	return
}
