package cmd

import (
	"os"

	"github.com/eezhee/eezhee/pkg/config"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// var cfgFile string

var (
	Verbose   bool
	AppConfig *config.AppConfig
	rootCmd   = &cobra.Command{
		Use:   "eezhee",
		Short: "Eezhee is a very simple way to deploy an app to a public cloud",
		Long: `Eezhee is a very simple way to deploy an app to a public cloud
								Complete documentation is available at http://github.com/eezhee/eezhee`,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
)

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// log.Fatal(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

}

func initConfig() {

	// find home directory
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	// setup config file
	viper.AddConfigPath(home)

	// setup logging
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	if Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// get app settings
	AppConfig := config.NewAppConfig()
	err = AppConfig.Load()
	if err != nil {
		log.Fatal(err)
	}

}
