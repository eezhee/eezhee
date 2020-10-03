package github

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

// MakeAPIRequest will get release info for the repo
func MakeAPIRequest(apiURL string) (data []byte, headers http.Header, err error) {

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
