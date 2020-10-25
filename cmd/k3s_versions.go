package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/eezhee/eezhee/pkg/k3s"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionsCmd)
}

var versionsCmd = &cobra.Command{
	Use:   "k3s_versions",
	Short: "List the versions of k3s that can be used",
	Long:  `Will check the k3s repo on github and get a list of all the releases available`,
	Run: func(cmd *cobra.Command, args []string) {
		err := getK3sVersions()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func getK3sVersions() error {

	k3sManager := k3s.NewManager()

	// get list of channels
	channels, err := k3sManager.Releases.GetChannels()
	if err != nil {
		return err
	}

	// print out results to user
	var releaseInfo k3s.Release

	// start with the pinned releases
	pinnedReleases := []string{"latest", "stable"}
	for _, channel := range pinnedReleases {

		channelInfo, err := k3sManager.Releases.GetChannel(channel)
		if err != nil {
			fmt.Println("invalid channel name")
		}
		releaseInfo.Parse(channelInfo.Latest)

		fmt.Printf("%s: ", channel)
		fmt.Printf(" %s", releaseInfo.Name)
		fmt.Printf("\n")
	}

	// print out the rest of the channels
	for _, channel := range channels {

		// ignore testing channels
		if strings.Contains(channel, "testing") {
			continue
		}
		// ignore 'stable' and 'latest'
		if channel[0:1] != "v" {
			continue
		}

		fmt.Printf("%s: ", channel)

		releases, err := k3sManager.Releases.GetReleases(channel)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, release := range releases {

			releaseInfo.Parse(release)
			fmt.Printf(" %s", releaseInfo.Name)
		}
		fmt.Printf("\n")

	}

	return nil
}
