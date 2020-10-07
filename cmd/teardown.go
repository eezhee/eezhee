package cmd

import (
	"fmt"
	"strings"

	"github.com/eezhee/eezhee/pkg/config"
	"github.com/eezhee/eezhee/pkg/digitalocean"
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

		teardownVM()
	},
}

// teardownVM will tear down the cluster & app
func teardownVM() {

	// see if there is a state file (so we know what we're supposed to teardown)
	deployStateFile := config.NewDeployState()
	if !deployStateFile.FileExists() {
		fmt.Println("app is not deployed so nothing to teardown")
		return
	}

	// load state file
	err := deployStateFile.Load()
	if err != nil {
		fmt.Println("error reading deploy state file")
		return
	}

	// see which cloud cluster created on
	cloud := deployStateFile.Cloud
	if strings.Compare(cloud, "digitalocean") != 0 {
		fmt.Println("state file reference cloud we don't support: ", cloud)
		return
	}

	// get details of VM
	ID := deployStateFile.ID
	if ID == 0 {
		fmt.Println("invalid VM ID:", ID, ".  Can no teardown VM")
		return
	}

	// ready to delete the cluster
	manager := digitalocean.NewManager()
	err = manager.DeleteVM(ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	deployStateFile.Delete()

}
