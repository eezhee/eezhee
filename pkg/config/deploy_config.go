package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// DeployConfig has details of how to deploy the cluster
// note: all these fields are optional
type DeployConfig struct {
	v            *viper.Viper // viper object
	Cloud        string       // which cloud cluster was create in
	Region       string       // where to deploy the cluster
	Name         string       // what to call the cluster
	K3sVersion   string       // version of k3s to use. ie: latest, stable, 1.18, 1.18.3
	Size         string       // VM size
	SSHPublicKey string       // which ssh key to allow to acces the VM(s)
}

// NewDeployConfig will create a new deploy file object
func NewDeployConfig() *DeployConfig {
	deploy := new(DeployConfig)

	name := "deploy"
	path := "."

	// set default filename
	deploy.v = viper.New()
	deploy.v.SetConfigName(name)
	deploy.v.AddConfigPath(path)
	deploy.v.SetConfigType("yaml")
	filename := path + string(os.PathSeparator) + name + ".yaml"
	deploy.v.SetConfigFile(filename)

	return deploy
}

// FileExists checks if a deploy config file exists in current directory
func (d *DeployConfig) FileExists() bool {

	// try and get info about file
	_, err := os.Stat(d.v.ConfigFileUsed())

	return err == nil
}

// Load a given deploy state file
func (d *DeployConfig) Load() error {

	if err := d.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn("nothing to teardown as no state file found")
		} else {
			log.Error("error reading deploy file: ", err)
		}
		return err
	}

	d.Name = d.v.GetString("name")
	d.Cloud = d.v.GetString("cloud")
	d.Region = d.v.GetString("region")
	d.K3sVersion = d.v.GetString("k3s-version")
	d.Size = d.v.GetString("size")
	d.SSHPublicKey = d.v.GetString("ssh-public-key")

	return nil
}

// Save details of a deploy to the deploy-state file
func (d *DeployConfig) Save() error {

	d.v.Set("cloud", d.Cloud)
	d.v.Set("name", d.Name)
	d.v.Set("region", d.Region)
	d.v.Set("size", d.Size)
	d.v.Set("ssh-public-key", d.SSHPublicKey)
	d.v.Set("k3s-version", d.K3sVersion)

	err := d.v.WriteConfig()
	if err != nil {
		log.Error("could not write deploy config to disk: ", err)
		return err
	}

	return nil
}

// Delete the deploy state file
func (d *DeployConfig) Delete() error {

	// remove deploy.yaml
	err := os.Remove(d.v.ConfigFileUsed())
	if err != nil {
		log.Error("could not delete deploy config file: ", err)
		return err
	}

	return nil
}
