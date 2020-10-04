package k3s

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/eezhee/eezhee/pkg/github"
)

// notes:
//   looks like the github graphql api needs an auth token, no matter which data you query
//   the REST API allows some the tags endpoint to be queried without an auth token

// TODO:
//    refactor getverion code
//    sort results into a map
//    have way to get latest, or latest for a version (ie 1.18)
//    use cases:
//      build latest version of k3s
//      build specific version of k3s
//      check if there is a newer version of a stream (ie 1.18)

//    run k3sup with --k3s-version

// Version of k3s
type Version struct {
	Track   string // ie. 1.19
	Major   string
	Minor   string
	Release string
}

// Manager will handle installation of k3s
type Manager struct {
	k3sReleases map[string][]string // list of available k3s versions, groups by track (ie 1.19)

	// we should cache versions. only update if call GetVersions
	// maybe we do the 'latest' trick to see if changed
	// http://github.com/rancher/k3s/releases/latest
}

// NewManager will create a new k3s manager
func NewManager() *Manager {
	m := new(Manager)

	// pre-fetch list of versions we can install
	m.GetVersions()

	return m
}

// parse a given version string into its components
func parseVersion(versionStr string) (version Version, err error) {

	// v1.19.2
	versionParts := strings.Split(versionStr, ".")

	version.Major = versionParts[0]
	version.Major = version.Major[1:len(version.Major)]
	version.Minor = versionParts[1]
	version.Release = versionParts[2]
	version.Track = version.Major + "." + version.Minor

	return version, err
}

// GetVersions of K3S that are available
func (m *Manager) GetVersions() (map[string][]string, error) {

	m.k3sReleases = make(map[string][]string)

	releases, err := github.GetRepoReleases("rancher", "k3s")
	if err != nil {
		return nil, err
	}

	// go through releases and filter out any non-final (non-RC) releases
	for _, release := range releases {

		// check tag name.  note, can't use 'releaseName' as is blank for older releases (pre-2020)
		tagName := release.TagName
		// releaseName := release.Name  		// older releases have this blank

		// parse release name
		fields := strings.Split(tagName, "+")
		// releaseParts[1] should always be 'k3s1'
		releaseParts := strings.Split(fields[0], "-")
		if len(releaseParts) == 1 {
			// only want final releases
			fullVersion := releaseParts[0]
			version, err := parseVersion(fullVersion)
			if err != nil {
				// skip this version
				fmt.Printf("could not parse version %s\n", fullVersion)
			}

			// sort into streams   1.16, 1.17, etc
			// ignore version before 1.16
			// note versions in each track will be in desending order (ie 1.19.2, 1.19.1)
			if strings.Compare(version.Track, "1.16") >= 0 {
				m.k3sReleases[version.Track] = append(m.k3sReleases[version.Track], fullVersion)
			}

		} else {
			// ignore non-final releases
			// fmt.Println("ignoring", releaseParts[0], releaseParts[1])
		}
	}

	return m.k3sReleases, nil
}

// CheckRequirements will make sure required components are installed
func (m *Manager) CheckRequirements() (bool, error) {

	if runtime.GOOS == "windows" {
		return false, errors.New("tool does not support windows yet")
	}

	// is brew installed?
	cmd := exec.Command("which", "brew")
	_, err := cmd.CombinedOutput()
	if err != nil {
		// brew not installed
		return false, errors.New("brew not installed. check https://brew.sh for instructions")
	}

	return true, nil
}

// setupK3sup will make sure k3sup on machine
func setupK3sup() bool {

	// see if k3sup installed
	cmd := exec.Command("which", "k3sup")
	_, err := cmd.CombinedOutput()
	if err != nil {

		// not installed, so let's try and install
		cmd := exec.Command("brew", "install", "k3sup")
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("could not install k3sup: %s\n", stdoutStderr)
			return false
		}

		// k3sup now installed
	}

	// get k3sup version
	cmd = exec.Command("k3sup", "version")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		// oopps, there is a k3sup but not working properly
		fmt.Println("k3sup not working properly.  Please reinstall")
		return false
	}
	k3supOutput := string(stdoutStderr)

	// parse and extract version
	var installedVersion = ""
	lines := strings.Split(k3supOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Version") {
			fields := strings.Split(line, ":")
			installedVersion = strings.TrimSpace(fields[1])
		}
	}
	if strings.Compare(installedVersion, "") == 0 {
		fmt.Println("could not get version of installed k3sup")
		return false
	}
	fmt.Printf("installed version: %s\n", installedVersion)

	// get current verison
	// request latest, github will redirect to specific version
	request, err := http.NewRequest("GET", "https://github.com/alexellis/k3sup/releases/latest", nil)
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	// defer response.Body.Close()
	if err != nil {
		fmt.Printf("could not get latest version of k3sup: %s\n", err)
		return false
	}
	finalURL := response.Request.URL.Path
	fields := strings.Split(finalURL, "/")
	latestVersion := fields[len(fields)-1]
	fmt.Printf("latest version: %s\n", latestVersion)

	// compare our version to latest
	result := strings.Compare(installedVersion, latestVersion)
	if result < 0 {
		fmt.Printf("newer version of k3sup (%s) available\n", latestVersion)
		fmt.Printf("using brew to upgrade k3sup\n")

		// this could take a while
		// brew upgrade k3sup
		cmd = exec.Command("brew", "upgrade", "k3sup")
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("brew could not upgrade k3sup: %s\n", stdoutStderr)
			return false
		}
		// now have latest k3sup

	} else if result > 0 {
		// our version is newer than what is on github
		// should never happen (so we will ignore)
	} else {
		// have the latest version
	}

	return true
}

// Install k3s using k3sup
func (m *Manager) Install() bool {

	// make sure we have brew
	_, err := m.CheckRequirements()
	if err != nil {
		fmt.Println(err)
		return false
	}

	if !setupK3sup() {
		// could not install/update k3sup
	}

	// now run k3sup bla bla
	fmt.Println("here is where we should run k3sup on the VM we created")

	// `k3sup install --ip $IP --ssh-key $KEY --user ubuntu`
	// need ip, name of ssh key, user

	return true
}
