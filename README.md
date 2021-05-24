# Eezhee

A super fast and easy way to create a k3s based kubernetes cluster on a variety of public clouds.  Currently DigitalOcean, Linode and Vultr are supported.  All it takes is a single command and about 2 minutes and your cluster is ready to use.  Most of the time is taken by the cloud provider bring up the base VM. Eezhee is ideal for development, testing or learning about Kubernetes.

What Eezhee does it combine the creation of a VM on the given provider and the installation of kubernetes.  It also does things like auto discover which is the closest region and what is the current stable version of kubernetes.  While you can just issue the `build` command and get a working cluster, you can also customize the cluster using a simple deploy file.  This allows you to specify which region or what version of kubernetes to install. 

This is early days for Eezhee.  There is lots more functionality I'd like to add.  See the `roadmap` below for details 

## Installation

Eezhee is written in Go so can be run on MacOS, Linux and Windows.  Currently most testing has been done on an x86 MacOS with just basic testing on the other OSs.  Please report any issues found with the other OSs. 

On MacOS, the easiest way to install eezhee is with brew.

```bash
brew tap eezhee/eezhee
brew install eezhee
```

If you want to update Eezhee, its as simple as:

```bash
brew update eezhee
```

To install on Windows or linux, grab the binaries from the [release section of github](https://github.com/eezhee/eezhee/releases).

## Configuration

If you have the cloud provider's CLI tool installed, then Eezhee will automatically discover your API KEY. Otherwise, you can use the `eezhee clouds {cloudname} {api_key}` command to set the API key Eezhee should use.  If you want to see which clouds are currently configured, type `eezhee clouds list`.   Note, you can config Eezhee to work with a single cloud or all the various supported clouds.

## Using Eezhee

### Create Kubernetes Cluster

All it takes to create a cluster is the build command.  

```bash
eezhee build
```

This will determine which cloud region is closest to you and create a single node k3s cluster using the most recent stable version of kubernetes.  It will also generate two files in the current directory.  First `.kubeconfig` file so you can use `kubectl` with your new cluster.  Eezhee will also create a `deploy-state.yaml` file with details about your cluster.

By default, your cluster will have the same name as your current directory.  This allows you to name your clusters to match your project names.

The default VM size of a cluster has 2GB of memory.  This currently can't be changed but should be in the next release (v0.3)

If you want customize how your cluster is built, create a `deploy.yaml` file with the settings.  It can just be a single setting (like the region to use) or several settings. See the `Deploy file` section and place the file in the current directory.  If you are using the cluster with a project, put the file in the projects root directory.

### Delete Cluster

When you no longer need your cluster, you can easily delete it with the `teardown` command.  Note, you need to be in same directory as the `build` command was run in as Eezhee looks for the `deploy-state.yaml` file to get details about the cluster.

```bash
eezhee teardown
```

### List Clusters

If you use Eezhee in several projects (or across multiple providers), it can be hard to keep track of how many clusters are running.  You can use the `list` command to see what is actually running on your cloud providers at any given point in time.  Eezhee uses tags to flag which VMs where created by eezhee.

```bash
eezhee list
```

### List Supported Kubernetes Versions

While the default it to build the cluster with the current stable version, many other version are also supported.  Use the `k3s_versions` command to list all the supported versions.   This will list each major version and each of the releases available for that version.

```bash
$ eezhee k3s_versions
latest:  v1.21.1
stable:  v1.20.7
v1.20:  v1.20.7 v1.20.6 v1.20.5 v1.20.4 v1.20.2 v1.20.0 v1.20.0
.
v1.16:  v1.16.15 v1.16.14 v1.16.13 v1.16.11 v1.16.10 v1.16.9 v1.16.7
```

## Help

To see all the commands Eezhee supports, type:

```bash
eezhee help
```

## Config Files

### Deploy Config File

Create a `deploy.yaml` file if you want to customize how the cluster is built.    The `build` command looks in the current directory for this file.  If found, the settings overide the defaults.

Currently you can set:

- `name`:  What you name your cluster. This will also be the kubectl context name. 
- `cloud`: Which provider to use.  This is only necessary if you have configured Eezhee to work with several providers
- `region`:  Defaults to the closest region.  You can set it to any of the providers regions. Note, right now regions name as provider specific but this is likely to change in the future as Eezhee config files should be provider agnostic.  
- `k3s-version`: Which version of Kubernetes to install.  It must be one reported with the `k3s_version` command.  Options include `stable`, `latest`, a channel (ie `v1.20`) or a specific version (ie `v1.20.3`)

### Deploy State File

Once a cluster has been created, Eezhee will create a `deploy-state.yaml` file in the current directory.  This has all the key details about your cluster.  This file should be considered read-only.

## Roadmap

Something things that are planned include:

- add more public clouds (aws is most likely next)
- ability to update the kubernetes version of a running cluster (within the same stream ie 1.20.1 -> 1.20.6)
- generic (provider agnostic) way of specifying a VM size or region
- ability to resize the VM your cluster uses
- ability to add nodes to your cluster
- support for multi-node clusters and using a variety of VM sizes

## Final Thoughts

Hopefully Eezhee will make it easier for you to leverage all the benefits of kubernetes without getting bogged down in the details.  

Good luck, have fun and let me know your thoughts!