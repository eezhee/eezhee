package cmd

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your running clusters",
	Long:  `All software has versions. This is Eezhee's`,
	Run: func(cmd *cobra.Command, args []string) {
		listVMs()
	},
}

func listVMs() {

	clouds := []string{"digitalocean", "linode", "vultr"}

	// // see which cloud we have an api token for
	// cloud := AppConfig.GetDefaultCloud()
	// if len(cloud) == 0 {
	// 	// opps, no api keys specified so can't proceed until resolved
	// 	log.Error("no cloud provider configured. User 'eezhee auth add'")
	// 	return false
	// }

	numClusters := 0

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
		}

		// go through all VMs and look for VMs that are tagged with 'eezhee'
		if len(vmInfo) > 0 {
			fmt.Println(cloud, ":")
			for i := range vmInfo {
				if len(vmInfo[i].Tags) > 0 {
					for _, tag := range vmInfo[i].Tags {
						if strings.Compare(tag, "eezhee") == 0 {
							// we created this VM
							createdTimestamp, err := time.Parse(time.RFC3339, vmInfo[i].CreatedAt)
							if err != nil {
								fmt.Println("error: ", err)
								continue
							}
							fmt.Printf("  %s (%s)  status: %s  created at: %s\n", vmInfo[i].Name, vmInfo[i].ID, vmInfo[i].Status,
								createdTimestamp.Format("2006-01-02 15:04"))
							numClusters = numClusters + 1
						}
					}
				}
			}
		}
	}

	if numClusters == 0 {
		fmt.Println("no clusters currently running")
	}

}
