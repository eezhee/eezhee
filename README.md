# Eezhee

A fast and easy way to create and manage kubernetes clusters on DigitalOcean.  A cluster can normally be built in 1-2 minutes.  Eezhee also makes it easy to update the version of kubernetes if there is a new release.

In the future, we plan to add support for resizing your cluster and adding base services like Let's Encrypt

## Status

Eezhee is still in the early days of being developed. But today it can build a single node k3s cluster with a single command.  Also, it supports `deploy.yaml` files which allow you to define the details of the cluster for each project.

## Installation

Currently Eezhee has only been tested on MacOS.  In the future, there will be support for Windows and Linux.

You can use brew to install Eezhee.  

```bash
brew tap eezhee/eezhee
brew install eezhee
```

If you want to update Eezhee, its as simple as:

```bash
brew update eezhee
```

Eezhee needs your PC to have DigitalOcean's CLI [`doctl`](https://www.digitalocean.com/docs/apis-clis/doctl/) installed and setup with an API token. If you have already been using DigitalOcean, you probably already have this done.  If not, you can find instructions [here](https://www.digitalocean.com/docs/apis-clis/doctl/how-to/install/)


## Usage

To see what commands Eezhee supports, type:

```bash
./eezhee help
```

### Create Kubernetes Cluster

This will create a single node k3s cluster in the region closest to the user and create a `kubectl` config file in the current directory.  By default, the most recent version of k3s will be used.

```bash
./eezhee build
```

### Delete Cluster

You can easily delete your cluster by using the `teardown` command.  Note, you need to be in the project direction as Eezhee looks for the `deploy-state.yaml` file to get details about the cluster.

```bash
./eezhee teardown
```

### List Clusters

If you use Eezhee in several projects, it can be hard to keep track of how many clusters are running.  You can use the `list` command to see what is actually running on DigitalOcean at any given point in time.

```bash
./eezhee list
```

Note, if you want to delete a given cluster, you need to switch to the corresponding project directory. See the note regarding the `deploy-state.yaml` file in the Delete command notes

### List Supported Kubernetes Versions 

If you want to configure your cluster to use a specific version of kubernetes, you can use the `list' command.  This will list each major version and each of the releases availalbe.

```bash
./eezhee k3s_versions
```

## Config Files

### Deploy Config File

When you create a cluster, Eezhee will look in the current directory to see if there is a `deploy.yaml` file.  This allows you to overide the defautls and set various parameters for your cluster.

Currently you can set:

- Size
- Region.  Defaults to the closest region.  You can set it to any of DigitalOceans regions.  The list of possible values includes: 

### Deploy State File

You can 
certain parameters, you can create a `deploy.yaml` file

// use the following to see details about the VM created
cat ./deploy-state.yaml
```

To delete the VM once you don't need it anymore, use:

```bash
./eezhee teardown
```

## Final Thoughts

Hopefully Eezhee will make it easier for you to leverage all the benefits of kubernetes without getting bogged down in the details.  
Good luck and have fun!
