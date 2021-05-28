package cmd

import (
	"fmt"

	"github.com/eezhee/eezhee/pkg/config"
	"github.com/eezhee/eezhee/pkg/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(defaultCmd)

	defaultCmd.AddCommand(listDefaultsCmd)
	defaultCmd.AddCommand(defaultProviderCmd)
	defaultCmd.AddCommand(defaultRegionCmd)
	defaultCmd.AddCommand(defaultVMSizeCmd)

}

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "list or set defaults",
	Long:  `list of set default values for Eezhee`,
}

// list app defaults
var listDefaultsCmd = &cobra.Command{
	Use:   "list",
	Short: "List app defaults",
	Long:  `List all the default settings for Eezhee"`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Defaults:")

		var value string
		appConfig := config.NewAppConfig()
		err := appConfig.Load()
		if err != nil {
			fmt.Println("error: could not load current settings")
			return
		}

		if len(appConfig.DefaultCloud) > 0 {
			value = appConfig.DefaultCloud
		} else {
			value = "not set"
		}
		fmt.Printf("  cloud: %s\n", value)

		if len(appConfig.DefaultVMSize) > 0 {
			value = appConfig.DefaultVMSize
		} else {
			value = "not set"
		}
		fmt.Printf("  vmsize: %s\n", value)

		if len(appConfig.DefaultRegion) > 0 {
			value = appConfig.DefaultRegion
		} else {
			value = "not set"
		}
		fmt.Printf("  region: %s\n", value)
	},
}

var defaultProviderCmd = &cobra.Command{
	Use:   "cloud",
	Short: "set cloud to use when building",
	Long:  `Set cloud to use for building when the config does not specify one`,
	Run: func(cmd *cobra.Command, args []string) {

		// get the input
		// make sure its a valid cloud we support
		defaultCloud := args[0]
		for _, cloud := range core.SupportedClouds {
			if defaultCloud == cloud {
				config := config.NewAppConfig()
				err := config.Load()
				if err != nil {
					fmt.Println("error: could not load config settings")
					return
				}
				config.DefaultCloud = defaultCloud
				err = config.Save()
				if err != nil {
					fmt.Println("error: could not save config settings")
				}
			}
		}
	},
}

var defaultVMSizeCmd = &cobra.Command{
	Use:   "vmsize",
	Short: "set size of VM to use when building",
	Long:  `Set size of VM  to use when building and the config does not specify one`,
	Run: func(cmd *cobra.Command, args []string) {

		// get the input
		defaultVMSize := args[0]
		fmt.Println(defaultVMSize)

		// TODO - is valid input
		config := config.NewAppConfig()
		err := config.Load()
		if err != nil {
			fmt.Println("error: could not load config settings")
			return
		}
		config.DefaultVMSize = defaultVMSize
		err = config.Save()
		if err != nil {
			fmt.Println("error: could not save config settings")
		}
	},
}

var defaultRegionCmd = &cobra.Command{
	Use:   "region",
	Short: "set region to build clusters on",
	Long:  `Set region to build on when the config does not specify one`,
	Run: func(cmd *cobra.Command, args []string) {

		// get the input
		defaultRegion := args[0]
		fmt.Println(defaultRegion)

		// TODO - is valid input
		config := config.NewAppConfig()
		err := config.Load()
		if err != nil {
			fmt.Println("error: could not load config settings")
			return
		}
		config.DefaultRegion = defaultRegion
		err = config.Save()
		if err != nil {
			fmt.Println("error: could not save config settings")
		}
	},
}
