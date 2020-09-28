package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the apps you have running",
	Long:  `All software has versions. This is Eezhee's`,
	Run: func(cmd *cobra.Command, args []string) {
		listVMs()
	},
}

func listVMs() error {

	// get a list of VMs running on DO
	cmd := exec.Command("doctl", "compute", "droplet", "list", "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return err
	}

	// parse the json output
	var vmInfo []digitalocean.VMInfo
	json.Unmarshal([]byte(stdoutStderr), &vmInfo)

	// go through all VMs and look for VMs that are tagged with 'eezhee'
	for i := range vmInfo {
		if len(vmInfo[i].Tags) > 0 {
			for _, tag := range vmInfo[i].Tags {
				if strings.Compare(tag, "eezhee") == 0 {
					// we created this VM
					fmt.Println(vmInfo[i].Name, " (", vmInfo[i].ID, ")  status:", vmInfo[i].Status, " created at:", vmInfo[i].CreatedAt)
				}
			}
		}
	}

	return nil
}
