package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// DeployState has details of the deploy-state file for a cluster
type DeployState struct {
	v            *viper.Viper // used to read/write state
	Cloud        string       // which cloud cluster was create in
	ID           string       // ID of the VM cluster is on
	Name         string       // name of the cluster
	Region       string       // region cluster deployed to
	Size         string       // VM size
	IP           string       // public IPv4 address
	SSHPublicKey string       // which ssh key authorited to access VM
	K3sVersion   string       // version of k3s installed
}

// NewDeployState will create a new deploy file object
func NewDeployState() (f *DeployState) {
	state := new(DeployState)

	name := "deploy-state"
	path := "."

	state.v = viper.New()
	state.v.SetConfigName(name)
	state.v.SetConfigType("yaml")
	state.v.AddConfigPath(path)
	filename := path + string(os.PathSeparator) + name + ".yaml"
	state.v.SetConfigFile(filename)
	// state.v.SetEnvPrefix("ez")

	return state
}

// FileExists checks if a deploy config file exists in current directory
func (s *DeployState) FileExists() bool {

	// try and get info about file
	_, err := os.Stat(s.v.ConfigFileUsed())

	return err == nil
}

// Load a given deploy state file
func (s *DeployState) Load() error {

	// try and read the file
	if err := s.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn("nothing to teardown as no state file found")
		} else {
			log.Error("error reading state file: ", err)
		}
		return err
	}

	s.Cloud = s.v.GetString("cloud")
	s.ID = s.v.GetString("id")
	s.Name = s.v.GetString("name")
	s.Region = s.v.GetString("region")
	s.Size = s.v.GetString("size")
	s.IP = s.v.GetString("ip")
	s.SSHPublicKey = s.v.GetString("ssh-public-key")
	s.K3sVersion = s.v.GetString("k3s-version")

	return nil
}

// Save details of a deploy to the deploy-state file
func (s *DeployState) Save() error {

	// move all our state variables over to viper
	s.v.Set("cloud", s.Cloud)
	s.v.Set("id", s.ID)
	s.v.Set("name", s.Name)
	s.v.Set("region", s.Region)
	s.v.Set("size", s.Size)
	s.v.Set("ip", s.IP)
	s.v.Set("ssh-public-key", s.SSHPublicKey)
	s.v.Set("k3s-version", s.K3sVersion)

	err := s.v.WriteConfig()
	if err != nil {
		log.Error("could not save state file: ", err)
		return err
	}

	return nil
}

// Delete the deploy state file
func (s *DeployState) Delete() error {

	// remove deploy.yaml
	err := os.Remove(s.v.ConfigFileUsed())
	if err != nil {
		log.Warn("could not remove state file: ", err)
		return err
	}

	return nil
}
