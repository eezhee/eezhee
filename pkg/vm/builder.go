package vm

import "github.com/eezhee/eezhee/pkg/config"

// BuildInfo has details of VM to create
type BuildInfo struct {
	deployConfig config.DeployConfig
	deployState  config.DeployState
}

// CreateVM will create a VM on the given cloud
func (b *BuildInfo) CreateVM() error {
	return nil
}
