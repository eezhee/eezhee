package cmd

import (
	"fmt"
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
		fmt.Println("building...done")

		buildVM()

	},
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
