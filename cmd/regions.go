package cmd

import (
	"fmt"

	"github.com/eezhee/eezhee/pkg/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(regionsCmd)
	regionsCmd.AddCommand(listRegionsCmd)
}

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "Details about the regions supported about cloud providers supported",
	Long:  `Details about the regions that Eezhee (and enabled providers) support`,
}

var listRegionsCmd = &cobra.Command{
	Use:   "list",
	Short: "list the available regions for each provider",
	Long:  `list the available regions for each provider`,
	Run: func(cmd *cobra.Command, args []string) {
		regionsList()
	},
}

func regionsList() {
	// for each provider that is configured
	// call GetRegions()
	// go through each cloud and see if enabled

	numEnabled := 0
	for _, cloud := range core.SupportedClouds {
		manager, err := GetManager(cloud)
		if err == nil {
			fmt.Println(cloud, ":")
			manager.GetRegions()
			numEnabled = numEnabled + 1
		}
	}

}
