package cmd

import (
	"fmt"
	"strings"

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

	clouds := []string{"digitalocean", "linode", "vultr"}

	// // see which cloud we have an api token for
	// cloud := AppConfig.GetDefaultCloud()
	// if len(cloud) == 0 {
	// 	// opps, no api keys specified so can't proceed until resolved
	// 	log.Error("no cloud provider configured. User 'eezhee auth add'")
	// 	return false
	// }

	for _, cloud := range clouds {

		manager, err := GetManager(cloud)
		if err != nil {
			// don't have api key for this cloud
			continue
		}

		// get all VMs in our account
		vmInfo, err := manager.ListVMs()
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

		if err != nil {
			log.Error(err)
		}

	}

	return true
}
