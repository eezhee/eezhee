package cmd

import (
	"errors"

	"github.com/eezhee/eezhee/pkg/aws"
	"github.com/eezhee/eezhee/pkg/config"
	"github.com/eezhee/eezhee/pkg/core"
	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/eezhee/eezhee/pkg/linode"
	"github.com/eezhee/eezhee/pkg/vultr"
)

// GetManager will create a new manager object for the desired public cloud
func GetManager(appConfig *config.AppConfig, cloud string) (vmManager core.VMManager, err error) {

	switch cloud {
	case "aws":
		// TODO: work out how to authenticate for aws
		vmManager = aws.NewManager(appConfig.LinodeAPIKey)
		if vmManager == nil {
			return nil, errors.New("could not create aws client")
		}
	case "digitalocean":
		vmManager = digitalocean.NewManager(appConfig.DigitalOceanAPIKey)
		if vmManager == nil {
			return nil, errors.New("could not create digitalocean client")
		}
	case "linode":
		vmManager = linode.NewManager(appConfig.LinodeAPIKey)
		if vmManager == nil {
			return nil, errors.New("could not create linode client")
		}
	case "vultr":
		vmManager = vultr.NewManager(appConfig.VultrAPIKey)
		if vmManager == nil {
			return nil, errors.New("could not create vultr client")
		}
	default:
		// should never get here (but lets play it safe)
		return nil, errors.New("invalid cloud type")
	}

	return vmManager, nil
}
