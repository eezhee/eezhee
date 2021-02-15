package k3d

import (
	"context"
	"log"

	k3dCluster "github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/config"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	"github.com/rancher/k3d/v4/pkg/runtimes/docker"
)

// CreateK3dCluster creates a local cluster
func CreateK3dCluster() {

	// create a config
	var name = "k3d-eezhee"

	// cfgViper is an empty viper config object
	var cfg config.SimpleConfig

	cfg.Name = name

	// work on SimpleConfig
	// then transform

	// SimpleConfig, ClusterConifg, ClusterListConfig??
	// SinpleConfig takes Viper config and unmarshals

	// servers - # of servers to create
	// agents - # of agents to create
	// image - which image to use for nodes
	// network - join existing network
	// token - cluster token
	// and lots more in k3d/cmd/cluster/clusterCreate.go

	cfg, err := config.FromViperSimple()

	// see if cluster exists??
	var ctx = context.Background()
	var runtime runtimes.Runtime = docker.Docker{}
	clusterConfig, err := config.TransformSimpleToClusterConfig(ctx, runtime, cfg)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := k3dCluster.ClusterGet(ctx, runtime, clusterConfig); err == nil {
		log.Fatalf("cluster %s already exists", name)
	}

	// create cluster

	// get kubeconfig

}

// DeleteK3dCluster will delete a k3d cluster
func DeleteK3dCluster() {

	k3dCluster.ClusterDelete()

}
