package k3s

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eezhee/eezhee/pkg/github"
	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

// notes:
//   looks like the github graphql api needs an auth token, no matter which data you query
//   the REST API allows some the tags endpoint to be queried without an auth token

// TODO:
//    refactor getverion code
//    sort results into a map
//    have way to get latest, or latest for a version (ie 1.18)
//    use cases:
//      build latest version of k3s
//      build specific version of k3s
//      check if there is a newer version of a stream (ie 1.18)

// Version of k3s
type Version struct {
	Track   string // ie. 1.19
	Major   string
	Minor   string
	Release string
}

// Manager will handle installation of k3s
type Manager struct {
	Releases map[string][]string // list of available k3s versions, groups by track (ie 1.19)

	// we should cache versions. only update if call GetVersions
	// maybe we do the 'latest' trick to see if changed
	// http://github.com/rancher/k3s/releases/latest
}

// NewManager will create a new k3s manager
func NewManager() *Manager {
	m := new(Manager)

	// pre-fetch list of versions we can install
	m.GetVersions()

	return m
}

// parse a given version string into its components
func parseVersion(versionStr string) (version Version, err error) {

	// v1.19.2
	versionParts := strings.Split(versionStr, ".")

	version.Major = versionParts[0]
	version.Major = version.Major[1:len(version.Major)]
	version.Minor = versionParts[1]
	version.Release = versionParts[2]
	version.Track = version.Major + "." + version.Minor

	return version, err
}

// GetVersions of K3S that are available
func (m *Manager) GetVersions() (map[string][]string, error) {

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
		fields := strings.Split(tagName, "+")
		// releaseParts[1] should always be 'k3s1'
		releaseParts := strings.Split(fields[0], "-")
		if len(releaseParts) == 1 {
			// only want final releases
			fullVersion := releaseParts[0]
			version, err := parseVersion(fullVersion)
			if err != nil {
				// skip this version
				fmt.Printf("could not parse version %s\n", fullVersion)
			}

			// sort into streams   1.16, 1.17, etc
			// ignore version before 1.16
			// note versions in each track will be in desending order (ie 1.19.2, 1.19.1)
			if strings.Compare(version.Track, "1.16") >= 0 {
				m.Releases[version.Track] = append(m.Releases[version.Track], fullVersion)
			}

		} else {
			// ignore non-final releases
			// fmt.Println("ignoring", releaseParts[0], releaseParts[1])
		}
	}

	return m.Releases, nil
}

// GetLatestVersion of k3s that can be installed
func (m *Manager) GetLatestVersion() (latestRelease string) {

	for track := range m.Releases {
		latestRelease = m.Releases[track][0]
	}

	return latestRelease
}

// CheckRequirements will make sure required components are installed
func (m *Manager) CheckRequirements() (bool, error) {

	if runtime.GOOS == "windows" {
		return false, errors.New("tool does not support windows yet")
	}

	// is brew installed?
	cmd := exec.Command("which", "brew")
	_, err := cmd.CombinedOutput()
	if err != nil {
		// brew not installed
		return false, errors.New("brew not installed. check https://brew.sh for instructions")
	}

	return true, nil
}

// Install k3s on given VM
func (m *Manager) Install(ipAddress string, k3sVersion string) bool {

	user := "root"
	sshPort := 22

	// build install command

	// `k3sup install --ip $IP --ssh-key $KEY --user ubuntu`
	// installk3sExec := fmt.Sprintf("INSTALL_K3S_EXEC='server %s --tls-san %s %s'", clusterStr, ipAddress, strings.TrimSpace(k3sExtraArgs))
	// installK3scommand := fmt.Sprintf("curl -sLS https://get.k3s.io | sh -\n", installk3sExec, k3sVersion)
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
	// fmt.Println(output)

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
	context := "appname"

	configUpdater := strings.NewReplacer(
		"127.0.0.1", ipAddress,
		"localhost", ipAddress,
		"default", context,
	)
	kubectlConfig := configUpdater.Replace(output)

	fmt.Println(kubectlConfig)

	// save output to kubectrl config file
	// TODO: option to merge?
	absPath, _ := filepath.Abs("kubeconfig")
	err = ioutil.WriteFile(absPath, []byte(kubectlConfig), 0600)
	if err != nil {
		fmt.Println(err)
		return false
	}

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
	fmt.Println(outputStr)

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
