package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// DeployFile has details of how to deploy the cluster
type DeployFile struct {
	Cloud string // which cloud cluster was create in
}

// Exists checks if a deploy config file exists in current directory
func (s *DeployFile) Exists() bool {
	return false
}

// Load a given deploy state file
func (s *DeployFile) Load() error {

	viper.SetConfigName("deploy")
	viper.SetConfigFile("./deploy.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("nothing to teardown as no state file found")
		} else {
			fmt.Println("error reading deploy file")
		}
		return err
	}

	return nil
}

// Save details of a deploy to the deploy-state file
func (s *DeployFile) Save() error {

	return nil
}

// Delete the deploy state file
func (s *DeployFile) Delete() error {

	// remove deploy.yaml
	err := os.Remove("./deploy-state.yaml")
	if err != nil {

		fmt.Println("could not remove deploy-state.yaml file")
		return err
	}

	return nil
}