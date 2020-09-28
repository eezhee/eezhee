package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/eezhee/eezhee/pkg/digitalocean"

	"github.com/go-ping/ping"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const maxPingTime = 750

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

// datacenters: ams2","ams3","blr1","fra1","lon1","nyc1","nyc2","nyc3","sfo1","sfo2","sfo3","sgp1","tor1"
// sizes:
//        "512mb","1gb","2gb","4gb","8gb","16gb","32gb","48gb","64gb",
//        "s-1vcpu-1gb","s-1vcpu-2gb","s-3vcpu-1gb","s-2vcpu-2gb","s-1vcpu-3gb","s-2vcpu-4gb","s-4vcpu-8gb","s-8vcpu-16gb","s-6vcpu-16gb",
//			  "s-8vcpu-32gb","s-12vcpu-48gb","s-16vcpu-64gb","s-20vcpu-96gb","s-24vcpu-128gb","s-32vcpu-192gb",
//        "m-16gb","m-32gb","m-64gb","m-128gb","m-224gb",
//        "m-1vcpu-8gb","m-2vcpu-16gb","m-4vcpu-32gb","m-8vcpu-64gb","m-16vcpu-128gb","m-24vcpu-192gb","m-32vcpu-256gb",
//				"m3-2vcpu-16gb","m3-4vcpu-32gb","m3-8vcpu-64gb","m3-16vcpu-128gb","m3-24vcpu-192gb","m3-32vcpu-256gb",
//				"m6-2vcpu-16gb","m6-4vcpu-32gb","m6-8vcpu-64gb","m6-16vcpu-128gb","m6-24vcpu-192gb","m6-32vcpu-256gb"
//        "c-2","c2-2vcpu-4gb","c-4","c2-4vpcu-8gb","c-8","c2-8vpcu-16gb","c-16","c2-16vcpu-32gb","c-32","c2-32vpcu-64gb",
// images:  no need as we always want to be on the same plain Ubuntu box

// TODO: the way to manage sizes is on a 3 (or more dimensional plane).  User decides what they want to increase and we figure out the right VM upgrade

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
	// fmt.Println(vmName)

	sshFingerprint, err := getSSHFingerprint()
	if err != nil {
		fmt.Println(err)
		return false
	}
	// fmt.Println(sshFingerprint)

	uploaded, err := checkSSHKeyUploaded(sshFingerprint)
	if !uploaded {
		fmt.Println(err)
		return false
	}

	region, err := selectClosestRegion()

	// get directory name
	imageName := "ubuntu-20-04-x64"
	vmSize := "s-1vcpu-1gb"

	// time to create the VM
	var vmInfo []digitalocean.VMInfo

	//TODO: add tags '--tag-names' such as 'eezhee' 'userName'
	cmd := exec.Command("doctl", "compute", "droplet", "create", vmName,
		"--image", imageName, "--size", vmSize, "--region", region, "--ssh-keys", sshFingerprint,
		"--tag-name", "eezhee", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return false
	}
	// fmt.Println(string(stdoutStderr))

	// parse the output
	// get the ID
	// save the ID
	json.Unmarshal([]byte(stdoutStderr), &vmInfo)

	// go through each key and see if it matches what is on this machine
	fmt.Println(vmInfo[0].ID)
	fmt.Println(vmInfo[0].Status)

	// save to a file using viper??
	// ipaddress
	// region
	// size

	// TODO:
	// return struct
	// figure out what to save and how

	return true
}

type sampleIPAddress struct {
	region    string
	ipAddress string
}

func getPingTime(ipAddress string) (pingTime int64, err error) {

	pinger, err := ping.NewPinger(ipAddress)
	pinger.Timeout = time.Millisecond * maxPingTime // milliseconds
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	pinger.Count = 5
	pinger.Run()                 // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats

	pingTime = stats.AvgRtt.Milliseconds()

	return pingTime, nil
}

type regionPingTimes struct {
	name      string
	ipAddress string
}

func selectClosestRegion() (string, error) {

	sampleIPs := []regionPingTimes{
		{"ams2", "206.189.240.1"},
		{"blr1", "143.110.180.2"},
		{"fra1", "138.68.109.1"},
		{"lon1", "209.97.176.1"},
		{"nyc1", "192.241.251.1"},
		{"sfo1", "198.199.113.1"},
		{"sgp1", "209.97.160.1"},
		{"tor1", "68.183.194.1"},
	}

	// default to NYC
	var bestRegion = "nyc1" // default to nyc

	// get ping time to each region
	// to see which is the closest
	var lowestPingTime = maxPingTime
	for _, region := range sampleIPs {
		pingTime, err := getPingTime(region.ipAddress)
		if err != nil {
			return "", err
		}
		// fmt.Println(region.name, ": ", pingTime, "mSec")

		// is this datacenter closer than others we've seen so far
		if int(pingTime) < lowestPingTime {
			bestRegion = region.name
			lowestPingTime = int(pingTime)
		}
	}

	return bestRegion, nil
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
