package cmd

import (
	"fmt"
	"strings"

	"github.com/eezhee/eezhee/pkg/config"
	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func teardownVM() {
	// see if there is a state file so we know what we're supposed to teardown

	deployFile := new(config.StateFile)

	// TODO really want to check for existance 1st

	_, err := deployFile.Load()
	if err != nil {
		fmt.Println("error reading state file")
		return
	}

	cloud := deployFile.Cloud

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("nothing to teardown as no state file found")
		} else {
			fmt.Println("error reading state file")
		}
		return
	}

	// see which cloud cluster created on
	cloud := deployFile.Cloud
	if strings.Compare(cloud, "digitalocean") != 0 {
		fmt.Println("state file reference cloud we don't support: ", cloud)
		return
	}

	manager := digitalocean.NewManager()

	// get details of VM
	ID := deployFile.ID
	if ID == 0 {
		fmt.Println("invalid VM ID:", ID, ".  Can no teardown VM")
		return
	}

	// ready to delete the cluster
	err := manager.DeleteVM(ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	deployFile.Delete()

}
