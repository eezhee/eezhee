package cmd

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "eezhee",
	Short: "Eezhee is a very simple way to deploy an app to a public cloud",
	Long: `Eezhee is a very simple way to deploy an app to a public cloud
							Complete documentation is available at http://github.com/eezhee/eezhee`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// log.Fatal(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {

	// setup config file
	// find home directory
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	viper.AddConfigPath(home)
}
