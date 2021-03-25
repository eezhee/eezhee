package cmd

import (
	"fmt"
	"strings"

	"github.com/eezhee/eezhee/pkg/config"
	log "github.com/sirupsen/logrus"
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

func listVMs() bool {

	// get app settings
	appConfig := config.NewAppConfig()
	err := appConfig.Load()
	if err != nil {
		return false
	}

	// see which cloud we have an api token for
	cloud := appConfig.GetDefaultCloud()
	if len(cloud) == 0 {
		// opps, no api keys specified so can't proceed until resolved
		log.Error("no cloud provider configured. User 'eezhee auth add'")
		return false
	}

	// make sure the cluster doesn't already exist
	// is there a deploy state file
	deployState := config.NewDeployState()
	if deployState.FileExists() {
		deployState.Load()
		cloud = deployState.Cloud
	}

	// create a manager for desired cloud
	vmManager, err := GetManager(appConfig, cloud)
	if err != nil {
		log.Error(err)
		return false
	}

	// get all VMs in our account
	vmInfo, err := vmManager.ListVMs()
	if err != nil {
		log.Error(err)
		return false
	}

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

	return true
}
