# Eezhee

A tool to make it easy to build production platforms and deploy an app

## Status

Still early days.  Currently the tool makes it easy to create a VM on DigitalOcean that can be used to load k3s onto.  Details of the VM are stored in the `deploy-state.yaml` file.  Finally, you can use the teardown command to easily delete the VM when you no longer need it

## Usage

To build a VM on digitalocean, use:

```bash
./eezhee build

// use the following to see details about the VM created
cat ./deploy-state.yaml
```

To delete the VM once you don't need it anymore, use:

```bash
./eezhee teardown
```

## Development

To build the binary

```bash
go build
```

To run the tests use the following

```bash
go test ./...
```
