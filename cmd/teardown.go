package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(teardownCmd)
}

var teardownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Will teardown a running app",
	Long:  `All software has versions. This is Eezhee's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tearing down... done")
		// read deploy.yaml
		// if no file, error out

		// get IP
		// try and find that VM with doctl tool
		// if no VM, error out

		// use doctl to delete the VM
		// remove deploy.yaml

	},
}
