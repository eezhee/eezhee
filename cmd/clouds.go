package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/eezhee/eezhee/pkg/core"
	"github.com/eezhee/eezhee/pkg/digitalocean"
	"github.com/eezhee/eezhee/pkg/linode"
	"github.com/eezhee/eezhee/pkg/vultr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cloudsCmd)
	cloudsCmd.AddCommand(listCloudsCmd)
	cloudsCmd.AddCommand(digitaloceanApiKeyCmd)
	cloudsCmd.AddCommand(linodeApiKeyCmd)
	cloudsCmd.AddCommand(vultrApiKeyCmd)
}

var cloudsCmd = &cobra.Command{
	Use:   "clouds",
	Short: "Details about cloud providers supported",
	Long:  `List or config cloud providers used by eezhee`,
}

// list which providers we support
var listCloudsCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported cloud providers",
	Long:  `List supported cloud providers and if they are configured"`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Enabled Clouds:")

		// get the subcommmand
		clouds := []string{"digitalocean", "linode", "vultr"}
		for _, cloud := range clouds {
			manager, err := GetManager(cloud)
			if err == nil {
				fmt.Println("  ", cloud)
				manager.ListVMs()
			}

			// else {
			// 	fmt.Println("  ", cloud, " (not configured)")
			// }
		}
	},
}

var digitaloceanApiKeyCmd = &cobra.Command{
	Use:     "digitalocean [api_key]",
	Aliases: []string{"do"},
	Short:   "Set digitalocean api key ",
	Long:    `Set digitalocean api key eezhee should use`,
	Args:    validateArguments,
	Run:     saveApiKey,
}

var linodeApiKeyCmd = &cobra.Command{
	Use:   "linode [api_key]",
	Short: "Set linode api key ",
	Long:  `Set linode api key eezhee should use`,
	Args:  validateArguments,
	Run:   saveApiKey,
}

var vultrApiKeyCmd = &cobra.Command{
	Use:   "vultr [api_key]",
	Short: "Set vultr api key ",
	Long:  `Set vultr api key eezhee should use`,
	Args:  validateArguments,
	Run:   saveApiKey,
}

// validateArguments will make sure api key is valid
func validateArguments(cmd *cobra.Command, args []string) (err error) {

	// make sure only one argument
	if len(args) < 1 {
		return errors.New("no api key provided")
	} else if len(args) > 1 {
		return errors.New("too many arguments specified. only an api key is required")
	}

	// get api key user provided
	apiKey := args[0]

	// need to know which cloud
	var manager core.VMManager
	if strings.HasPrefix(cmd.Use, "digitalocean") || strings.HasPrefix(cmd.Use, "do") {
		manager, err = digitalocean.NewManager(apiKey)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(cmd.Use, "linode") {
		manager, err = linode.NewManager(apiKey)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(cmd.Use, "vultr") {
		manager, err = vultr.NewManager(apiKey)
		if err != nil {
			return err
		}
	} else {
		// cobra will make sure this never is allowed
		// ie this code should never be called
		log.Error("invalid cloud name")
	}

	// validate api key
	_, err = manager.ListVMs()
	if err != nil {
		return errors.New("invalid api key specified")
	}

	return nil
}

// save the api key provided
func saveApiKey(cmd *cobra.Command, args []string) {

	// get api key user provided
	apiKey := args[0]

	// store api key in app config
	// need to know which cloud
	if strings.HasPrefix(cmd.Use, "digitalocean") || strings.HasPrefix(cmd.Use, "do") {
		AppConfig.DigitalOceanAPIKey = apiKey
	} else if strings.HasPrefix(cmd.Use, "linode") {
		AppConfig.LinodeAPIKey = apiKey
	} else if strings.HasPrefix(cmd.Use, "vultr") {
		AppConfig.VultrAPIKey = apiKey
	} else {
		// cobra will make sure this never is allowed
		// ie this code should never be called
		log.Error("invalid cloud name")
	}

	// now save to disk
	err := AppConfig.Save()
	if err != nil {
		log.Error("could not save api key to config file")
		return
	}

	log.Info("api key added to config")
}
