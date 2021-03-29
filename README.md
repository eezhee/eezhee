# Eezhee

A super fast and easy way to create a k3s based kubernetes cluster on a variety of public clouds.  Currently DigitalOcean, Linode and Vultr are support.  All it takes is a single command and about 2 minutes and your cluster is ready to use.  Eezhee is normally used for development, testing or for learning about kubernetes.

Eezhee will auto discover which cloud you are using and which is the closest region. That means all you have to do is install Eezhee and issue a single `build` command and you will have working cluster.  Or with a simple deploy file, you can set which cloud you want to use and which region to build it in. A given eezhee deploy files will across all providers so you can easily move from one cloud to another.

By default, Eezhee will install the most recent stable version of kubernetes but you can set it to install a wide variety of versions.

## Installation

Eezhee is written in go so can be run on MacOS, Linux and Windows.  Currently most testing has been done on MacOS so please report any issues with the other OSs.

You can use brew to install Eezhee.  

```bash
brew tap eezhee/eezhee
brew install eezhee
```

If you want to update Eezhee, its as simple as:

```bash
brew update eezhee
```

To install on Windows or linux, grab the binaries from the release section of github.

## Configuring

If you have the cloud provider's CLI tool installed, then eezhee will automatically discover your API KEY.  Otherwise, you can use the `eezhee cloud {cloudname} {api_key}` command to set

You can config eezhee to work with a single or all the various supported clouds.  If you want to see which clouds are currently configured, use the `eezhee clouds list` command.

## Usage

To see what commands Eezhee supports, type:

```bash
./eezhee -h
```

### Create Kubernetes Cluster

you can create a cluster simply with the build command

```bash
./eezhee build
```

This will create a single node k3s cluster in the region closest to the user and create a `kubectl` config file in the current directory.  By default, the most recent version of k3s will be used.  

If you want to specify which version of kubernetes you want installed or which region to use, see the `Deploy file` section and put the file in the current directory.

### Delete Cluster

You can easily delete your cluster by using the `teardown` command.  Note, you need to be in same directory as the `build` command was run in as Eezhee looks for the `deploy-state.yaml` file to get details about the cluster.

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

- Name.  What you name your cluster. This also matches the kubectl context name
- Region.  Defaults to the closest region.  You can set it to any of DigitalOceans regions.  The list of possible values includes:

### Deploy State File

Once a cluster has been created, Eezhee will create a `deploy-state.yaml` file in the current directory.  This has all the key details about your cluster.  This file should be considered read-only.

## Roadmap

The following items are scheduled to be added to eezhee

- add more public clouds (aws is most likely next)
- support for multi-node clusters and using a variety of VM sizes
- ability to update or upgrade the kubernetes version of your cluster

## Final Thoughts

Hopefully Eezhee will make it easier for you to leverage all the benefits of kubernetes without getting bogged down in the details.  

Good luck and have fun!
