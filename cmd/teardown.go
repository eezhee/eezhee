package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eezhee/eezhee/pkg/config"
	"github.com/eezhee/eezhee/pkg/core"
	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/eezhee/eezhee/pkg/linode"
	"github.com/eezhee/eezhee/pkg/vultr"
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
		return errors.New("app is not deployed. nothing to teardown")
	}

	// load state file
	err = deployStateFile.Load()
	if err != nil {
		return errors.New("error reading deploy state file")
	}

	// see which cloud cluster created on
	cloud := deployStateFile.Cloud
	var manager core.VMManager

	switch cloud {
	case "digitalocean":
		manager = digitalocean.NewManager(appConfig.DigitalOceanAPIKey)
	case "linode":
		manager = linode.NewManager(appConfig.LinodeAPIKey)
	case "vultr":
		manager = vultr.NewManager(appConfig.VultrAPIKey)
	default:
		return errors.New("state file reference cloud we don't support: ")
	}

	// get details of VM
	ID := deployStateFile.ID
	if ID == 0 {
		msg := fmt.Sprintf("invalid VM ID: %d - Can not teardown VM\n", ID)
		return errors.New(msg)
	}

	// ready to delete the cluster
	err = manager.DeleteVM(ID)
	if err != nil {
		return err
	}
	fmt.Println("k3s cluster (and VM) deleted")

	// remove the kubeconfig file
	kubeConfigFile, _ := filepath.Abs("kubeconfig")
	_ = os.Remove(kubeConfigFile)
	// if err == nil {
	// 	fmt.Println("removed kubeconfig for cluster")
	// }

	// remove the deploy state file
	err = deployStateFile.Delete()
	if err != nil {
		return err
	}
	// fmt.Println("deploy-state file deleted")

	return nil
}
