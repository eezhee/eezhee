package cmd

import (
	"fmt"
	"sort"

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
		getK3sVersions()
	},
}

func getK3sVersions() error {

	k3sManager := k3s.NewManager()

	// get the possible versions
	releaseVersions, err := k3sManager.GetVersions()
	if err != nil {
		return err
	}

	// note, since we are using a map, tracks are not sorted
	// get a list of tracks
	// sort
	// use that to print results in decreasing version #
	var tracks []string
	for track := range releaseVersions {
		tracks = append(tracks, track)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tracks)))

	// print out results to user
	for _, track := range tracks {
		fmt.Printf("%s: ", track)
		for _, version := range releaseVersions[track] {
			fmt.Printf(" %s", version)
		}
		fmt.Printf("\n")
	}

	return nil
}
