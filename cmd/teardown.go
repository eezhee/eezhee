package cmd

import (
	"errors"
	"fmt"
	"os"
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

		err := teardownVM()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// teardownVM will tear down the cluster & app
func teardownVM() error {

	// get app settings
	appConfig := config.NewAppConfig()
	err := appConfig.Load()
	if err != nil {
		return err
	}

	// see if there is a state file (so we know what we're supposed to teardown)
	deployStateFile := config.NewDeployState()
	if !deployStateFile.FileExists() {
		return errors.New("app is not deployed so nothing to teardown")
	}

	// load state file
	err = deployStateFile.Load()
	if err != nil {
		return errors.New("error reading deploy state file")
	}

	// see which cloud cluster created on
	cloud := deployStateFile.Cloud
	if strings.Compare(cloud, "digitalocean") != 0 {
		return errors.New("state file reference cloud we don't support: ")
	}

	// get details of VM
	ID := deployStateFile.ID
	if ID == 0 {

		msg := fmt.Sprintf("invalid VM ID: %d - Can not teardown VM\n", ID)
		return errors.New(msg)
	}

	// ready to delete the cluster
	manager := digitalocean.NewManager(appConfig.DigitalOceanAPIKey)
	err = manager.DeleteVM(ID)
	if err != nil {
		return err
	}

	err = deployStateFile.Delete()
	if err != nil {
		return err
	}

	return nil
}
