package k3s

// code to manage k3s release channels and their associated releases
// note, by default we don't load all the releases as it itsn't that fast
// by default, we have a list of all the release channels and what the associated
// latest release is

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/eezhee/eezhee/pkg/github"
)

// Channel has info about a specific k3s release channel
type Channel struct {
	Name   string `json:"name"`   // id and name normally the same
	Latest string `json:"latest"` // ie v1.18.10+k3s1
	// LatestRegex  string `json:"latestRegexp"`
	// ExcludeRegex string `json:"excludeRegexp"`
}

// Release of k3s
type Release struct {
	FullName   string // v1.19-rc1+k3s1
	Name       string // v1.19-rc1
	Channel    string // 1.19
	Major      string // 1
	Minor      string // 19
	Patch      string // 2
	Extra      string // rc1
	K3sRelease string // k3s1
}

// ReleaseInfo is a list of k3s release channels
type ReleaseInfo struct {
	Channels []Channel           `json:"data"`
	Releases map[string][]string // list of available k3s versions, groups by track (ie 1.19)
}

// Parse will take a given version string into parse into its components
func (r *Release) Parse(releaseStr string) (err error) {

	// currently support version in the following format:
	// v1.19.2-rc1+k3s2
	// v1.19.2

	if len(releaseStr) == 0 {
		return err
	}

	r.FullName = releaseStr

	// see if has k3s release info ( +k3s1)
	parts := strings.Split(r.FullName, "+")
	if len(parts) > 1 {
		r.K3sRelease = parts[1]
	}
	r.Name = parts[0]

	// now split
	parts = strings.Split(r.Name, ".")

	r.Major = parts[0][1:len(parts[0])]
	r.Minor = parts[1]
	r.Channel = "v" + r.Major + "." + r.Minor

	parts = strings.Split(parts[2], "-")
	r.Patch = parts[0]
	if len(parts) > 1 {
		r.Extra = parts[1]
	}

	return err
}

// LoadChannels will get the channel details from updates.k3s.io
func (ri *ReleaseInfo) LoadChannels() error {

	// setup api request
	apiURL := k3sUpdateAPI + k3sChannelsEndpoint
	request, err := http.NewRequest("GET", apiURL, nil)
	request.Header.Add("Accept", "application/json")

	// make the api request
	client := &http.Client{Timeout: apiTimeout}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// get channel data
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// parse the data
	err = json.Unmarshal([]byte(data), ri)
	if err != nil {
		return err
	}

	return nil
}

// LoadReleases will get a list of all
func (ri *ReleaseInfo) LoadReleases() error {

	// see if we already have the versions list
	if ri.Releases != nil {
		return nil
	}

	ri.Releases = make(map[string][]string)

	githubReleases, err := github.GetRepoReleases("rancher", "k3s")
	if err != nil {
		return err
	}

	// go through releases and filter out any non-final (non-RC) releases
	for _, githubRelease := range githubReleases {

		// ignore release if it is in draft stage
		if githubRelease.Draft {
			continue
		}

		// if release.Prerelease {
		// 	continue
		// }

		// check tag name.  note, can't use 'releaseName' as is blank for older releases (pre-2020)
		tagName := githubRelease.TagName
		// releaseName := githubRelease.Name  		// older releases have this blank

		// parse release name
		var release Release
		release.Parse(tagName)
		if len(release.K3sRelease) == 0 {
			// needs to have +k3sx in release name
			continue
		}

		// check if RC build
		if len(release.Extra) > 0 {
			// extra not empty so not normal build
			continue
		}

		// sort into streams   1.16, 1.17, etc
		// ignore version before 1.16
		// note versions in each track will be in desending order (ie 1.19.2, 1.19.1)
		if strings.Compare(release.Channel, "1.16") >= 0 {
			ri.Releases[release.Channel] = append(ri.Releases[release.Channel], release.FullName)
		}

	}

	return nil
}

// GetChannel info on desired release channel
func (ri *ReleaseInfo) GetChannel(desiredChannel string) (channel Channel, err error) {

	for _, channel = range ri.Channels {
		if strings.Compare(channel.Name, desiredChannel) == 0 {
			return channel, nil
		}
	}
	return channel, errors.New("invalid channel name")
}

// GetChannels returns array of all valid channel names
func (ri *ReleaseInfo) GetChannels() (channels []string, err error) {

	// go throgh all channels and build a list of all their names
	for _, channel := range ri.Channels {
		channels = append(channels, channel.Name)
	}

	// now sort in descending order
	sort.Sort(sort.Reverse(sort.StringSlice(channels)))

	return channels, nil
}

// GetLatestRelease of k3s that is available for a given channel
func (ri *ReleaseInfo) GetLatestRelease(desiredChannel string) (latestRelease string, err error) {

	// is channel valid
	channel, err := ri.GetChannel(desiredChannel)
	if err != nil {
		return "", errors.New("invalid channel name")
	}

	// get a list of tracks to see which is the
	latestRelease = channel.Latest

	return latestRelease, nil
}

// GetReleases will return all the releases for a given channel
func (ri *ReleaseInfo) GetReleases(desiredChannel string) (releases []string, err error) {

	if len(desiredChannel) == 0 {
		return nil, errors.New("no channel specified")
	}

	if releases, ok := ri.Releases[desiredChannel]; ok {
		return releases, nil
	}

	return nil, errors.New("invalid channel name")
}
