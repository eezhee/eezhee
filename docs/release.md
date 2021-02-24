# Building the Code

## Old Method

1. update the version number in `/cmd/version.go`
1. build the binary with `go build`
1. create sha with `sha256sum eezhee > eezhee-sha256sum.sha`
1. create a draft release on github.  add details of what is in the release
1. upload the binary and signature to the draft release
1. publish the release

## Creating a Test Build (v2)

1. build the binaries with `goreleaser release --rm-dist --skip-upload`

This will create the binaries and associated files in `dist`.  Review the changelog 

## Creating a Release (v2)

1. update the version number in `/cmd/version.go`.  Check in the change and merge into master.
1. tag the commit with `git tag vX.Y.Z` and `git push origin vX.Y.Z`
1. make sure you have the GITHUB_TOKEN environment variable setup
1. run `goreleaser` on your laptop.  `goreleaser release --rm-dist`

## Cleanup

1. remove `release` dir & `eezhee-sha256sum.sha`
