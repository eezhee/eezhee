package config

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// AppConfig has details of how to deploy the cluster
// note: all these fields are optional
type AppConfig struct {
	v                  *viper.Viper // viper object
	DigitalOceanAPIKey string
	LinodeAPIKey       string
	VultrAPIKey        string
	CloudFlareAPIKey   string
	DefaultCloud       string // required if we have more than one api key
	DefaultVMSize      string
	DefaultRegion      string
}

// NewAppConfig will create a new deploy file object
func NewAppConfig() *AppConfig {
	config := new(AppConfig)

	// config will be stored in .eezhee subdir of users home directory
	// homedir package is cross platform so works on all common OSs
	homeDir, _ := homedir.Dir()
	path := homeDir + string(os.PathSeparator) + ".eezhee"

	// make sure directory exists
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Error("could not create config directory ", path)
		log.Error(err)
		return nil
	}

	// make sure file can be read or if no file, new file created
	name := "config"
	filename := path + string(os.PathSeparator) + name + ".yaml"

	file, err := os.OpenFile(filename, os.O_CREATE, 0755)
	if err != nil {
		log.Error("could not create config file. ", err)
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

// Load a given deploy state file
func (a *AppConfig) Load() error {

	if err := a.v.ReadInConfig(); err != nil {
		log.Error("could not read config file: ", err)
		return err
	}

	a.DigitalOceanAPIKey = a.v.GetString("digitalocean-api-key")
	a.LinodeAPIKey = a.v.GetString("linode-api-key")
	a.VultrAPIKey = a.v.GetString("vultr-api-key")
	a.CloudFlareAPIKey = a.v.GetString("cloudflare-api-key")

	a.DefaultCloud = a.v.GetString("default-cloud")
	a.DefaultVMSize = a.v.GetString("default-vmsize")
	a.DefaultRegion = a.v.GetString("default-region")

	return nil
}

// Save details of a deploy to the deploy-state file
func (a *AppConfig) Save() error {

	a.v.Set("digitalocean-api-key", a.DigitalOceanAPIKey)
	a.v.Set("linode-api-key", a.LinodeAPIKey)
	a.v.Set("vultr-api-key", a.VultrAPIKey)
	a.v.Set("cloudflare-api-key", a.CloudFlareAPIKey)

	a.v.Set("default-cloud", a.DefaultCloud)
	a.v.Set("default-vmsize", a.DefaultVMSize)
	a.v.Set("default-region", a.DefaultRegion)

	err := a.v.WriteConfig()
	if err != nil {
		log.Error("could not save config to disk: ", err)
		return err
	}

	return nil
}

// GetDefaultCloud returns which cloud has been configured
// if there are multiple, will default to DigitalOcean
func (a *AppConfig) GetDefaultCloud() string {

	if len(a.DefaultCloud) > 0 {
		return a.DefaultCloud
	} else {
		// no default set so just return 1st that has API key
		if len(a.DigitalOceanAPIKey) > 0 {
			return "digitalocean"
		} else if len(a.LinodeAPIKey) > 0 {
			return "linode"
		} else if len(a.VultrAPIKey) > 0 {
			return "vultr"
		}
	}

	// don't have any API keys so can't have a default
	log.Debug("no cloud has been configured")
	return ""
}
