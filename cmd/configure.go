package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(digitaloceanApiKeyCmd)
	// configCmd.AddCommand(linodeApiKeyCmd)
	// configCmd.AddCommand(vultrApiKeyCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure eezhee settings",
	Long:  `Configue eezhee app settings such as api key to use for a given cloud`,
	Run: func(cmd *cobra.Command, args []string) {

		err := configure()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	},
}

// teardownVM will tear down the cluster & app
func configure() error {

	// get the subcommmand

	return nil
}

var digitaloceanApiKeyCmd = &cobra.Command{
	Use:   "digitalocean",
	Short: "Configure eezhee settings",
	Long:  `Configue eezhee app settings such as api key to use for a given cloud`,
	Run: func(cmd *cobra.Command, args []string) {

		// err := configure()
		// if err != nil {
		// 	log.Error(err)
		// 	os.Exit(1)
		// }
	},
}
