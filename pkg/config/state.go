package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// StateFile has details of the deploy-state file for a cluster
type StateFile struct {
	Cloud string // which cloud cluster was create in
	ID    int    // ID of the VM cluster is on
}

// Exists checks if a deploy-state file exists in current directory
func (s *StateFile) Exists() bool {
	return false
}

// Load a given deploy state file
func (s *StateFile) Load() error {

	viper.SetConfigName("deploy-state")
	viper.SetConfigFile("./deploy-state.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("nothing to teardown as no state file found")
		} else {
			fmt.Println("error reading state file")
		}
		return err
	}

	return nil
}

// Save details of a deploy to the deploy-state file
func (s *StateFile) Save() error {

	return nil
}

// Delete the deploy state file
func (s *StateFile) Delete() error {

	// remove deploy.yaml
	err := os.Remove("./deploy-state.yaml")
	if err != nil {

		fmt.Println("could not remove deploy-state.yaml file")
		return err
	}

	return nil
}
