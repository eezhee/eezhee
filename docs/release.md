# Creating a Release

1. update the version number in `/cmd/version.go`
1. build the binary with `go build`
1. create sha with `sha256sum eezhee > eezhee-sha256sum.sha`
1. create a draft release on github.  add details of what is in the release
1. upload the binary and signature to the draft release
1. publish the release
