package config

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// AppConfig has details of how to deploy the cluster
// note: all these fields are optional
type AppConfig struct {
	v                  *viper.Viper // viper object
	DigitalOceanAPIKey string
	CloudFlareAPIKey   string
	LinodeAPIKey       string
}

// NewAppConfig will create a new deploy file object
func NewAppConfig() *AppConfig {
	config := new(AppConfig)

	name := "config"
	homeDir, _ := homedir.Dir()
	path := homeDir + "/.eezhee"
	filename := path + "/" + name + ".yaml"

	// make sure directory exists
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Println("could not create directory ~/.eezhee for config file")
		return nil
	}

	// make sure file can be read or if no file, new file created
	file, err := os.OpenFile(filename, os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("could not initalize config file in directory ~/.eezhee")
		return nil
	}
	file.Close()

	// set default filename
	config.v = viper.New()
	config.v.SetConfigName(name)
	config.v.AddConfigPath(path)
	config.v.SetConfigType("yaml")
	config.v.SetConfigFile(filename)

	return config
}

// FileExists checks if a deploy config file exists in current directory
func (a *AppConfig) FileExists() bool {

	// try and get info about file
	_, err := os.Stat(a.v.ConfigFileUsed())

	return err == nil
}

// Load a given deploy state file
func (a *AppConfig) Load() error {

	if err := a.v.ReadInConfig(); err != nil {
		fmt.Println("could not read eezhee config file: ", err)
		return err
	}

	a.DigitalOceanAPIKey = a.v.GetString("digitalocean-api-key")
	a.CloudFlareAPIKey = a.v.GetString("cloudflare-api-key")
	a.LinodeAPIKey = a.v.GetString("linode-api-key")

	return nil
}

// Save details of a deploy to the deploy-state file
func (a *AppConfig) Save() error {

	if len(a.DigitalOceanAPIKey) > 0 {
		a.v.Set("digitalocean-api-key", a.DigitalOceanAPIKey)
	}
	if len(a.CloudFlareAPIKey) > 0 {
		a.v.Set("cloudflare-api-key", a.CloudFlareAPIKey)
	}
	if len(a.LinodeAPIKey) > 0 {
		a.v.Set("linode-api-key", a.CloudFlareAPIKey)
	}

	err := a.v.WriteConfig()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Delete the deploy state file
func (a *AppConfig) Delete() error {

	// remove deploy.yaml
	err := os.Remove(a.v.ConfigFileUsed())
	if err != nil {

		fmt.Printf("could not remove %s file\n", a.v.ConfigFileUsed())
		return err
	}

	return nil
}
