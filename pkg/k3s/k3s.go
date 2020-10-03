package k3s

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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/eezhee/eezhee/pkg/github"
)

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

func getVersionUsingREST() []string {

	var k3sReleases []string

	apiURL := "https://api.github.com/repos/rancher/k3s/releases?page=1&per_page=100"

	for len(apiURL) > 0 {
		data, headers, err := github.MakeAPIRequest(apiURL)
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
		var releases []github.Release

		json.Unmarshal([]byte(data), &releases)
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

			} else {
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
func GetVersions() []string {
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
