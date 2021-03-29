# Eezhee

A super fast and easy way to create a k3s based kubernetes cluster on a variety of public clouds.  Currently DigitalOcean, Linode and Vultr are supported.  All it takes is a single command and about 2 minutes and your cluster is ready to use.  Eezhee is normally used for development, testing or for learning about kubernetes.

Eezhee will auto discover which cloud you are using and which is the closest region. That means all you have to do is install Eezhee and issue a single `build` command and you will have a working cluster. Or you can create a simple deploy file if you want to set which cloud to use, which region to build in or which version of kubernetes to install. Deploy files are generally cloud-agnostic so if you change cloud providers, not much has to change.

By default, Eezhee will install the most recent stable version of kubernetes but you can set it to install a wide variety of kubernetes versions. You can use the `eezhee k3s_versions` to see which versions are available.

## Installation

Eezhee is written in go so can be run on MacOS, Linux and Windows.  Currently most testing has been done on an x86 MacOS so please report any issues found with the other OSs.  I have begun testing Windows and Linux and will update this README with more details shortly.

The easiest way to install eezhee is with brew.

```bash
brew tap eezhee/eezhee
brew install eezhee
```

If you want to update Eezhee, its as simple as:

```bash
brew update eezhee
```

To install on Windows or linux, grab the binaries from the [release section of github](https://github.com/eezhee/eezhee/releases).

## Configuring

If you have the cloud provider's CLI tool installed, then eezhee will automatically discover your API KEY. You can see which clouds eezhee has found using the `eezhee cloud list` command.  If your cloud is not listed, you will need to set the api key using the `eezhee cloud {cloudname} {api_key}` command.

Note, you can config eezhee to work with a single cloud or all the various supported clouds.

## Usage

To see what commands Eezhee supports, type:

```bash
eezhee -h
```

### Create Kubernetes Cluster

All it takes to create a cluster is the build command.  

```bash
eezhee build
```

This will determine which cloud region is closest to you and create a single node k3s cluster using the most recent stable version of kubernetes.  It will also generate two files in the current directory.  First `.kubeconfig` file so you can use `kubectl` with your new cluster.  Eezhee will also create a `deploy-state.yaml` file with details about your cluster.

If you want change any of these default, all you have to do is create a simple deploy file. This is where you specify which settings you want to override. It can just be a single setting (like the region to use) or several settings. See the `Deploy file` section and put the file in the current directory.

### Delete Cluster

You can easily delete your cluster by using the `teardown` command.  Note, you need to be in same directory as the `build` command was run in as Eezhee looks for the `deploy-state.yaml` file to get details about the cluster.

```bash
eezhee teardown
```

### List Clusters

If you use Eezhee in several projects, it can be hard to keep track of how many clusters are running.  You can use the `list` command to see what is actually running on your cloud providerat any given point in time.  Eezhee uses tags to flag which VMs where created by eezhee.

```bash
eezhee list
```

Note, if you want to delete a given cluster, you need to switch to the corresponding project directory. See the note regarding the `deploy-state.yaml` file in the Delete command notes

### List Supported Kubernetes Versions

If you want to configure your cluster to use a specific version of kubernetes, you can use the `list' command.  This will list each major version and each of the releases availalbe.

```bash
eezhee k3s_versions
```

You can then put the desired version into a `deploy.yaml` file that you use with the `build` command.

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
