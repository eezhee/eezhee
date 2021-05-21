package cmd

import (
	"errors"

	"github.com/eezhee/eezhee/pkg/aws"
	"github.com/eezhee/eezhee/pkg/core"
	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/eezhee/eezhee/pkg/linode"
	"github.com/eezhee/eezhee/pkg/vultr"
)

// GetManager will create a new manager object for the desired public cloud
func GetManager(cloud string) (core.VMManager, error) {

	var vmManager core.VMManager

	switch cloud {
	case "aws":
		// TODO: work out how to authenticate for aws
		vmManager, err := aws.NewManager(AppConfig.LinodeAPIKey)
		if err != nil {
			return vmManager, errors.New("could not create aws client")
		}
	case "digitalocean":
		vmManager, err := digitalocean.NewManager(AppConfig.DigitalOceanAPIKey)
		if err != nil {
			return vmManager, errors.New("could not create digitalocean client")
		}
	case "linode":
		vmManager, err := linode.NewManager(AppConfig.LinodeAPIKey)
		if err != nil {
			return vmManager, errors.New("could not create linode client")
		}
	case "vultr":
		vmManager, err := vultr.NewManager(AppConfig.VultrAPIKey)
		if err != nil {
			return vmManager, errors.New("could not create vultr client")
		}
	default:
		// should never get here (but lets play it safe)
		return vmManager, errors.New("invalid cloud type")
	}

	return vmManager, nil
}
