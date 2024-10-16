package k3s

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	k3sUpdateAPI        = "https://update.k3s.io"
	k3sChannelsEndpoint = "/v1-release/channels"
	apiTimeout          = 10 * time.Second
)

const maxRetries = 6               // number of times to try ssh'ing into the VM
const retryDelay = 5 * time.Second // time between retries

// use cases:
//  	build latest version of k3s
//		build specific version of k3s
// 		check if there is a newer version of a stream (ie 1.18)

// Manager will handle installation of k3s
type Manager struct {
	Releases ReleaseInfo
}

// NewManager will create a new k3s manager
func NewManager() *Manager {
	m := new(Manager)

	err := m.Releases.LoadChannels()
	if err != nil {
		log.Error("error: could not load list of k3s channels")
	}
	err = m.Releases.LoadReleases()
	if err != nil {
		log.Error("could not load list of k3s releases")
	}

	return m
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
	// log.Debug(installK3scommand)

	// get the private sshkey
	// TODO: should support ssh agent & passphrases
	keyFile := "~/.ssh/id_rsa"
	sshPrivateKeyFile, _ := homedir.Expand(keyFile)
	passphrase := ""

	signer, err := getSSHKey(sshPrivateKeyFile, passphrase)
	if err != nil {
		log.Error(err)
		return false
	}

	// setup ssh details
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	address := fmt.Sprintf("%s:%d", ipAddress, sshPort)

	// ssh into the server (& retry if can't)
	numRetries := 0
	conn := &ssh.Client{}
	for numRetries < maxRetries {

		// try and ssh into vm
		conn, err = ssh.Dial("tcp", address, config)
		if err == nil {
			// able to ssh into vm
			break
		}

		log.Debug(err)

		// wait a few seconds
		time.Sleep(retryDelay)
		numRetries += 1
	}

	if err != nil {
		return false
	}

	// install k3s on the VM
	output, err := runCommand(conn, installK3scommand)
	if err != nil {
		log.Error(err)
		log.Debug(output)
		return false
	}
	log.Info("k3s installed on VM")

	// get kubectl config
	getK3sConfigCommand := "cat /etc/rancher/k3s/k3s.yaml\n"
	output, err = runCommand(conn, getK3sConfigCommand)
	if err != nil {
		log.Error(err)
		log.Debug(string(output))
		return false
	}
	// log.Debug(string(output))

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
	err = os.WriteFile(absPath, []byte(kubectlConfig), 0600)
	if err != nil {
		log.Error(err)
		return false
	}

	log.Info("kubernetes is initializing")
	log.Info("you can access using `kubectl --kubeconfig ./kubeconfig get pods`")

	return true
}

func runCommand(conn *ssh.Client, command string) (outputStr string, err error) {

	sess, err := conn.NewSession()
	if err != nil {
		log.Error(err)
		return "", err
	}
	defer sess.Close()

	output, err := sess.CombinedOutput(command)
	outputStr = string(output)
	if err != nil {
		log.Error(err)
		log.Debug(outputStr)
		return outputStr, err
	}
	// log.Debug(outputStr)

	return outputStr, nil
}

// getSshKey
func getSSHKey(keyFilename string, passphrase string) (signer ssh.Signer, err error) {

	// load the private key
	privateKey, err := os.ReadFile(keyFilename)
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
