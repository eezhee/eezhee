package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eezhee/eezhee/pkg/config"
	log "github.com/sirupsen/logrus"
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
			log.Error(err)
			os.Exit(1)
		}
	},
}

// teardownVM will tear down the cluster & app
func teardownVM() error {

	// see if there is a state file (so we know what we're supposed to teardown)
	deployStateFile := config.NewDeployState()
	if !deployStateFile.FileExists() {
		return errors.New("app is not deployed. nothing to teardown")
	}

	// load state file
	err := deployStateFile.Load()
	if err != nil {
		return errors.New("error reading deploy state file")
	}

	// see which cloud cluster created on
	cloud := deployStateFile.Cloud

	// create a manager for desired cloud
	vmManager, err := GetManager(cloud)
	if err != nil {
		log.Error(err)
		return err
	}

	// get details of VM
	ID := deployStateFile.ID
	if len(ID) == 0 {
		msg := fmt.Sprintf("invalid VM ID: %s - Can not teardown VM\n", ID)
		return errors.New(msg)
	}

	// prompt to make sure user really wants to do this
	response := ""
	fmt.Printf("are you sure you want to delete %s cluster (Y/n)? ", deployStateFile.Name)
	fmt.Scanln(&response)
	response = strings.ToLower(response)
	if (response != "") && (response != "y") {
		return errors.New("deletion aborted")
	}

	// ready to delete the cluster
	err = vmManager.DeleteVM(ID)
	if err != nil {
		return err
	}
	log.Info("k3s cluster (and VM) deleted")

	// remove the kubeconfig file
	kubeConfigFile, _ := filepath.Abs("kubeconfig")
	_ = os.Remove(kubeConfigFile)
	if err == nil {
		log.Debug("removed kubeconfig for cluster")
	}

	// remove the deploy state file
	err = deployStateFile.Delete()
	if err != nil {
		return err
	}
	log.Debug("deploy-state file deleted")

	return nil
}
