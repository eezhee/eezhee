package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// DeployConfig has details of how to deploy the cluster
// note: all these fields are optional
type DeployConfig struct {
	v      *viper.Viper // viper object
	Cloud  string       // which cloud cluster was create in
	Name   string       // what to call the cluster
	Region string       // where to deploy the cluster
	Size   string       // VM size
}

// NewDeployConfig will create a new deploy file object
func NewDeployConfig() *DeployConfig {
	deploy := new(DeployConfig)

	name := "deploy"
	path := "./"

	// set default filename
	deploy.v = viper.New()
	deploy.v.SetConfigName(name)
	deploy.v.AddConfigPath(path)
	deploy.v.SetConfigType("yaml")
	filename := path + name + ".yaml"
	deploy.v.SetConfigFile(filename)

	return deploy
}

// FileExists checks if a deploy config file exists in current directory
func (d *DeployConfig) FileExists() bool {

	// try and get info about file
	_, err := os.Stat(d.v.ConfigFileUsed())
	if err != nil {
		return false
	}

	return true
}

// Load a given deploy state file
func (d *DeployConfig) Load() error {

	if err := d.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("nothing to teardown as no state file found")
		} else {
			fmt.Println("error reading deploy file")
		}
		return err
	}

	d.Cloud = d.v.GetString("cloud")
	d.Name = d.v.GetString("name")
	d.Region = d.v.GetString("region")
	d.Size = d.v.GetString("size")

	return nil
}

// Save details of a deploy to the deploy-state file
func (d *DeployConfig) Save() error {

	d.v.Set("cloud", d.Cloud)
	d.v.Set("name", d.Name)
	d.v.Set("region", d.Region)
	d.v.Set("size", d.Size)

	err := d.v.WriteConfig()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Delete the deploy state file
func (d *DeployConfig) Delete() error {

	// remove deploy.yaml
	err := os.Remove(d.v.ConfigFileUsed())
	if err != nil {

		fmt.Printf("could not remove %s file\n", d.v.ConfigFileUsed())
		return err
	}

	return nil
}
