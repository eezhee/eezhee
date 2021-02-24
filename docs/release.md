# Building the Code

## Old Method

1. update the version number in `/cmd/version.go`
1. build the binary with `go build`
1. create sha with `sha256sum eezhee > eezhee-sha256sum.sha`
1. create a draft release on github.  add details of what is in the release
1. upload the binary and signature to the draft release
1. publish the release

## Creating a Build (v2)

1. build the binaries with `goreleaser release --rm-dist --snapshot`

## Creating a Release (v2)

1. update the version number in `/cmd/version.go`
1. tag the commit with `git tag vX.Y.Z`
1. when you merge to master, a new release will be created automatically

## Cleanup

1. remove `release` dir & `eezhee-sha256sum.sha`
