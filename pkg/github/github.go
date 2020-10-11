package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Commit has basic commit info: url and sha
type Commit struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}

// Tags has details of a tag in a repo
type Tags struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     Commit `json:"commit"`
	NodeID     string `json:"node_id"`
}

// Asset details a file assosociated with a release
type Asset struct {
	URL                string `json:"url"`
	ID                 int64  `json:"id"`
	NodeID             string `json:"node_id"`
	Name               string `json:"name"`
	Label              string `json:"label"`
	Uploader           Author `json:"uploader"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               int64  `json:"size"`
	DownloadCount      int64  `json:"download_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Author has details about the author of a commit
type Author struct {
	Login             string `json:"login"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

// Release has details of a release for a repo
type Release struct {
	URL             string  `json:"url"`
	AssetsURL       string  `json:"assets_url"`
	UploadURL       string  `json:"upload_url"`
	HTMLURL         string  `json:"html_url"`
	ID              int64   `json:"id"`
	NodeID          string  `json:"node_id"`
	TagName         string  `json:"tag_name"`
	TargetCommitish string  `json:"target_commitish"`
	Name            string  `json:"name"`
	Draft           bool    `json:"draft"`
	Author          Author  `json:"author"`
	Prerelease      bool    `json:"prerelease"`
	CreatedAt       string  `json:"created_at"`
	PublishedAt     string  `json:"published_at"`
	Assets          []Asset `json:"assets"`
}

// GetRepoReleases will return a list of all the releases for a given repo
func GetRepoReleases(owner string, repo string) (repoReleases []Release, err error) {

	// get all releases, 100 at a time
	// paginate through results if more than 100
	apiURL := "https://api.github.com/repos/" + owner + "/" + repo + "/releases?page=1&per_page=100"
	for len(apiURL) > 0 {
		data, headers, err := makeRepoReleasesRequest(apiURL)
		if err != nil {
			return nil, err
		}

		// following caching headers are available: ETag & Cache-Control
		// following rate limiting headers are available: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset

		// extract the release data
		// format: array of 'release' structs
		// e.g. v1.17.1-alpha1+k3s1  (note releases before 2020 have different format)

		// extract the release data
		var releases []Release
		err = json.Unmarshal([]byte(data), &releases)
		if err != nil {
			return repoReleases, err
		}

		// add each release to array that will be returned
		repoReleases = append(repoReleases, releases...)

		// reset API URL - will only set again if there are more pages to load
		apiURL = ""

		// see if there are more releases
		linkHeader := headers.Get("Link")
		links := strings.Split(linkHeader, ",")
		for _, linkInfo := range links {
			linkFields := strings.Split(linkInfo, ";")
			link := linkFields[0]
			linkType := strings.TrimSpace(linkFields[1])

			if strings.Compare(linkType, "rel=\"next\"") == 0 {
				apiURL = strings.Trim(link, "<>")
			}
		}

	}

	return repoReleases, err
}

// GetVersionUsingREST will get the versions of k3s available on github
func GetVersionUsingREST() []string {

	var k3sReleases []string

	apiURL := "https://api.github.com/repos/rancher/k3s/releases?page=1&per_page=100"

	for len(apiURL) > 0 {
		data, headers, err := makeRepoReleasesRequest(apiURL)
		if err != nil {
			return nil
		}

		// following headers are available if we need them
		// header := response.Header.Get("ETag")
		// header = response.Header.Get("Cache-Control")
		// header = response.Header.Get("Bla")
		// header = response.Header.Get("X-Github-Media-Type")
		// header = response.Header.Get("X-RateLimit-Limit")
		// header = response.Header.Get("X-RateLimit-Remaining")
		// header = response.Header.Get("X-RateLimit-Reset")

		// extract the release data
		// format: v1.17.1-alpha1+k3s1  (note releases before 2020 have different format)
		var releases []Release
		err = json.Unmarshal([]byte(data), &releases)
		if err != nil {
			return nil
		}

		fmt.Println(len(releases))
		for _, release := range releases {
			tagName := release.TagName
			// releaseName := release.Name  		// older releases have this blank

			// parse release name
			fields := strings.Split(tagName, "+")
			// releaseParts[1] should always be 'k3s1'
			releaseParts := strings.Split(fields[0], "-")
			if len(releaseParts) == 1 {
				// only want final releases

				// ideally want to sort into streams   1.16, 1.17, etc
				// split by major.minor
				k3sReleases = append(k3sReleases, releaseParts[0])

				// } else {
				// ignore non-final releases
				// fmt.Println("ignoring", releaseParts[0], releaseParts[1])
			}
		}

		apiURL = ""

		// see if there are more releases
		linkHeader := headers.Get("Link")
		links := strings.Split(linkHeader, ",")
		for _, linkInfo := range links {
			linkFields := strings.Split(linkInfo, ";")
			link := linkFields[0]
			linkType := strings.TrimSpace(linkFields[1])

			if strings.Compare(linkType, "rel=\"next\"") == 0 {
				fmt.Println(link)
				apiURL = strings.Trim(link, "<>")
			}
		}

	}

	// return an map with 'major release' as key
	return k3sReleases
}

// GetVersions will return a list of k3s versions that can be downloaded
// func getVersionsUsingGraphQL() bool {

// 	// need to use github api
// 	// note, we will be unauthenticated and have a rate limit of 60/hour

// 	jsonData := map[string]string{
// 		"query": `
// 					{
// 						rateLimit {
// 							limit
// 							cost
// 							remaining
// 							resetAt
// 						}
// 						repository(owner:"rancher", name: "k3s") {
// 							refs(refPrefix: "refs/tags/", last: 1) {
// 								nodes {
// 									repository {
// 										releases(first:1, orderBy: {field: CREATED_AT, direction: DESC}) {
// 											nodes {
// 												name
// 												createdAt
// 												url
// 												releaseAssets(last: 4) {
// 													nodes {
// 														name
// 														downloadCount
// 														downloadUrl
// 													}
// 												}
// 											}
// 										}
// 									}
// 								}
// 							}
// 						}
// 					}
// 			`,
// 	}
// 	jsonValue, _ := json.Marshal(jsonData)
// 	request, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
// 	client := &http.Client{Timeout: time.Second * 10}
// 	response, err := client.Do(request)
// 	if err != nil {
// 		fmt.Printf("The HTTP request failed with error %s\n", err)
// 	}
// 	defer response.Body.Close()
// 	data, _ := ioutil.ReadAll(response.Body)
// 	fmt.Println(string(data))

// 	// need parse json that came back
// 	fmt.Println("Please format response")

// 	return true
// }

// makeRepoReleasesRequest will get release info for the repo
func makeRepoReleasesRequest(apiURL string) (data []byte, headers http.Header, err error) {

	// setup api request
	request, err := http.NewRequest("GET", apiURL, nil)
	request.Header.Add("User-agent", "eezhee")
	request.Header.Add("Accept", "application/vnd.github.v3+json")

	// make the api request
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return data, headers, err
	}
	defer response.Body.Close()

	// get relase data
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return data, headers, err
	}

	// get headers
	headers = response.Header

	return data, headers, err
}
