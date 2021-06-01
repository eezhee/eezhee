package cmd

import (
	"fmt"

	"github.com/eezhee/eezhee/pkg/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sizesCmd)
	sizesCmd.AddCommand(listSizesCmd)
}

var sizesCmd = &cobra.Command{
	Use:   "sizes",
	Short: "Details about VM sizes support",
	Long:  `Details about the vm sizes support by enabled cloud providers`,
}

var listSizesCmd = &cobra.Command{
	Use:   "list",
	Short: "list the available VM sizes for each provider",
	Long:  `list the available VM sizes for each provider`,
	Run: func(cmd *cobra.Command, args []string) {
		sizesList()
	},
}

func sizesList() {
	// for each provider that is configured
	// call GetRegions()
	// go through each cloud and see if enabled

	numEnabled := 0
	for _, cloud := range core.SupportedClouds {
		manager, err := GetManager(cloud)
		if err == nil {
			fmt.Println(cloud, ":")
			manager.GetVMSizes()
			numEnabled = numEnabled + 1
		}
	}

}
