package k3s

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/eezhee/eezhee/pkg/github"
	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

const (
	k3sUpdateAPI        = "https://update.k3s.io"
	k3sChannelsEndpoint = "/v1-release/channels"
	apiTimeout          = 10 * time.Second
)

// use cases:
//  	build latest version of k3s
//		build specific version of k3s
// 		check if there is a newer version of a stream (ie 1.18)

// Manager will handle installation of k3s
type Manager struct {
	Releases      map[string][]string // list of available k3s versions, groups by track (ie 1.19)
	StableChannel string
	LatestChannel string
}

// NewManager will create a new k3s manager
func NewManager() *Manager {
	m := new(Manager)

	return m
}

// GetVersions of K3S that are available
func (m *Manager) GetVersions() (map[string][]string, error) {

	// see if we already have the versions list
	if m.Releases != nil {
		return m.Releases, nil
	}

	m.Releases = make(map[string][]string)

	releases, err := github.GetRepoReleases("rancher", "k3s")
	if err != nil {
		return nil, err
	}

	// go through releases and filter out any non-final (non-RC) releases
	for _, release := range releases {

		// check tag name.  note, can't use 'releaseName' as is blank for older releases (pre-2020)
		tagName := release.TagName
		// releaseName := release.Name  		// older releases have this blank

		// parse release name
		var version Release
		fields := strings.Split(tagName, "+")
		// releaseParts[1] should always be 'k3s1'
		releaseParts := strings.Split(fields[0], "-")
		if len(releaseParts) == 1 {
			// only want final releases
			fullVersion := releaseParts[0]
			err := version.Parse(fullVersion)
			if err != nil {
				// skip this version
				fmt.Printf("could not parse version %s\n", fullVersion)
			}

			// sort into streams   1.16, 1.17, etc
			// ignore version before 1.16
			// note versions in each track will be in desending order (ie 1.19.2, 1.19.1)
			if strings.Compare(version.Channel, "1.16") >= 0 {
				m.Releases[version.Channel] = append(m.Releases[version.Channel], fullVersion)
			}

			// } else {
			// ignore non-final releases
			// fmt.Println("ignoring", releaseParts[0], releaseParts[1])
		}
	}

	return m.Releases, nil
}

// GetChannels returns array of all valid channel names
func (m *Manager) GetChannels() (channels []string, err error) {

	// make sure we have the list of versions
	if m.Releases == nil {
		_, err := m.GetVersions()
		if err != nil {
			return channels, err
		}
	}

	// go throgh all channels and build a list of all their names
	for channel := range m.Releases {
		channels = append(channels, channel)
	}

	// now sort in descending order
	sort.Sort(sort.Reverse(sort.StringSlice(channels)))

	return channels, nil
}

// GetLatestVersion of k3s that is available for a given channel
func (m *Manager) GetLatestVersion(channel string) (latestRelease string, err error) {

	// is channel valid
	_, exists := m.Releases[channel]
	if !exists {
		return "", errors.New("invalid channel name")
	}

	// get a list of tracks to see which is the
	latestRelease = m.Releases[channel][0]

	return latestRelease, nil
}

// CheckRequirements will make sure required components are installed
func (m *Manager) CheckRequirements() (bool, error) {

	if runtime.GOOS == "windows" {
		return false, errors.New("tool does not support windows yet")
	}

	return true, nil
}

// Install k3s on given VM
func (m *Manager) Install(ipAddress string, k3sVersion string, appName string) bool {

	user := "root"
	sshPort := 22

	// build install command
	installK3scommand := fmt.Sprintf("curl -sLS https://get.k3s.io | INSTALL_K3S_VERSION=%s sh -\n", k3sVersion)
	// fmt.Println(installK3scommand)

	// get the private sshkey
	// TODO: should support ssh agent & passphrases
	keyFile := "~/.ssh/id_rsa"
	sshPrivateKeyFile, _ := homedir.Expand(keyFile)
	passphrase := ""

	signer, err := getSSHKey(sshPrivateKeyFile, passphrase)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// connect to the server
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	address := fmt.Sprintf("%s:%d", ipAddress, sshPort)
	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// install k3s on the VM
	output, err := runCommand(conn, installK3scommand)
	if err != nil {
		fmt.Println(err)
		fmt.Println(output)
		return false
	}
	fmt.Println("k3s installed on VM")

	// get kubectl config
	getK3sConfigCommand := "cat /etc/rancher/k3s/k3s.yaml\n"
	output, err = runCommand(conn, getK3sConfigCommand)
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(output))
		return false
	}
	// fmt.Println(string(output))

	// need to update kubeconfig so works outside the VM
	// IP address needs to be set to external IP
	// also rename context to something other than 'default'
	context := appName

	configUpdater := strings.NewReplacer(
		"127.0.0.1", ipAddress,
		"localhost", ipAddress,
		"default", context,
	)
	kubectlConfig := configUpdater.Replace(output)

	// save output to kubectrl config file
	// TODO: option to merge?
	absPath, _ := filepath.Abs("kubeconfig")
	err = ioutil.WriteFile(absPath, []byte(kubectlConfig), 0600)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("cluster is ready")
	fmt.Println("you can access using `kubectl --kubeconfig .\\kubeconfig get pods`")

	return true
}

func runCommand(conn *ssh.Client, command string) (outputStr string, err error) {

	sess, err := conn.NewSession()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer sess.Close()

	output, err := sess.CombinedOutput(command)
	outputStr = string(output)
	if err != nil {
		fmt.Println(err)
		fmt.Println(outputStr)
		return outputStr, err
	}
	// fmt.Println(outputStr)

	return outputStr, nil
}

// getSshKey
func getSSHKey(keyFilename string, passphrase string) (signer ssh.Signer, err error) {

	// load the private key
	privateKey, err := ioutil.ReadFile(keyFilename)
	if err != nil {
		return nil, err
	}

	// decode the key
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(privateKey)
	}

	return signer, err
}
