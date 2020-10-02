package k3s

// notes:
//   looks like the github graphql api needs an auth token, no matter which data you query
//    the REST API allows some the tags endpoint to be queried without an auth token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// GithubCommit has basic commit info: url and sha
type GithubCommit struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}

// GithubTags has details of a tag in a repo
type GithubTags struct {
	Name       string       `json:"name"`
	ZipballURL string       `json:"zipball_url"`
	TarballURL string       `json:"tarball_url"`
	Commit     GithubCommit `json:"commit"`
	NodeID     string       `json:"node_id"`
}

// GithubAsset details a file assosociated with a release
type GithubAsset struct {
	URL                string       `json:"url"`
	ID                 int64        `json:"id"`
	NodeID             string       `json:"node_id"`
	Name               string       `json:"name"`
	Label              string       `json:"label"`
	Uploader           GithubAuthor `json:"uploader"`
	ContentType        string       `json:"content_type"`
	State              string       `json:"state"`
	Size               int64        `json:"size"`
	DownloadCount      int64        `json:"download_count"`
	CreatedAt          string       `json:"created_at"`
	UpdatedAt          string       `json:"updated_at"`
	BrowserDownloadURL string       `json:"browser_download_url"`
}

// GithubAuthor has details about the author of a commit
type GithubAuthor struct {
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

// GithubRelease has details of a release for a repo
type GithubRelease struct {
	URL             string        `json:"url"`
	AssetsURL       string        `json:"assets_url"`
	UploadURL       string        `json:"upload_url"`
	HTMLURL         string        `json:"html_url"`
	ID              int64         `json:"id"`
	NodeID          string        `json:"node_id"`
	TagName         string        `json:"tag_name"`
	TargetCommitish string        `json:"target_commitish"`
	Name            string        `json:"name"`
	Draft           bool          `json:"draft"`
	Author          GithubAuthor  `json:"author"`
	Prerelease      bool          `json:"prerelease"`
	CreatedAt       string        `json:"created_at"`
	PublishedAt     string        `json:"published_at"`
	Assets          []GithubAsset `json:"assets"`
}

func checkRequirements() bool {
	// see if k3sup installed
	// see if k3sup is the latest version
	// can we install it (brew install k3sup)
	return true
}

// install k3s using k3sup
func install() bool {
	return true
}

func getVersionUsingREST() bool {
	// curl -H "User-agent: eezhee" https://api.github.com/repos/rancher/k3s/releases

	// TODO: need to set User-Agent and Accept??

	request, err := http.NewRequest("GET", "https://api.github.com/repos/rancher/k3s/releases", nil)
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(data))

	// json parse
	var releases []GithubRelease

	json.Unmarshal([]byte(data), &releases)

	fmt.Println(len(releases))
	fmt.Println(releases[0])
	return true
}

// GetVersions will return a list of k3s versions that can be downloaded
func getVersionsUsingGraphQL() bool {

	// need to use github api
	// note, we will be unauthenticated and have a rate limit of 60/hour

	jsonData := map[string]string{
		"query": `
					{ 
						rateLimit {
							limit
							cost
							remaining
							resetAt
						}					
						repository(owner:"rancher", name: "k3s") {
							refs(refPrefix: "refs/tags/", last: 1) {
								nodes {
									repository {
										releases(first:1, orderBy: {field: CREATED_AT, direction: DESC}) {
											nodes {
												name
												createdAt
												url
												releaseAssets(last: 4) {
													nodes {
														name
														downloadCount
														downloadUrl
													}
												}	
											}
										}		
									}
								}
							}
						}
					}
			`,
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(data))

	// need parse json that came back
	fmt.Println("Please format response")

	return true
}

// GetVersions of K3S that are available
func GetVersions() bool {
	return getVersionUsingREST()
}

// {
//   repository(owner: "rancher", name: "k3s") {
//     refs(refPrefix: "refs/tags/", last: 10) {
//       nodes {
//         repository {
//           releases(first: 10, orderBy: {field: CREATED_AT, direction: DESC}) {
//             nodes {
//               name
//               createdAt
//               url
//               releaseAssets(last: 10) {
//                 nodes {
//                   name
//                   downloadCount
//                   downloadUrl
//                 }
//               }
//             }
//           }
//         }
//       }
//     }
//   }
// }
