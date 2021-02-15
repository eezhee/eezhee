package k3d

import (
	"context"
	"fmt"
	"os"
	"runtime"

	k3dCluster "github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/config"
	conf "github.com/rancher/k3d/v4/pkg/config/v1alpha2"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	k3d "github.com/rancher/k3d/v4/pkg/types"
	log "github.com/sirupsen/logrus"
)

// CreateK3dCluster creates a local cluster
func CreateK3dCluster() {

	// create a config
	var cfg conf.SimpleConfig

	cfg.Name = k3d.DefaultClusterName
	// cfg.Name = "k3d-eezhee"
	cfg.Image = k3d.DefaultK3sImageRepo + ":latest"
	// cfg.Image = "docker.io/rancher/k3s:latest" //TODO make try and get a list like we do with k3s binaries

	// servers - # of servers to create
	// agents - # of agents to create
	// image - which image to use for nodes
	// network - join existing network
	// token - cluster token
	// and lots more in k3d/cmd/cluster/clusterCreate.go

	// take simple config and add everything else we need (ie which runtime to use)
	var ctx = context.Background()
	// var runtime runtimes.Runtime = docker.Docker{}
	clusterConfig, err := config.TransformSimpleToClusterConfig(ctx, runtimes.SelectedRuntime, cfg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Debugf("===== Merged Cluster Config =====\n%+v\n===== ===== =====\n", clusterConfig)

	clusterConfig, err = config.ProcessClusterConfig(*clusterConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Debugf("===== Processed Cluster Config =====\n%+v\n===== ===== =====\n", clusterConfig)

	if err := config.ValidateClusterConfig(ctx, runtimes.SelectedRuntime, *clusterConfig); err != nil {
		log.Fatalln("Failed Cluster Configuration Validation: ", err)
	}

	// see if cluster exists??
	if _, err := k3dCluster.ClusterGet(ctx, runtimes.SelectedRuntime, &clusterConfig.Cluster); err == nil {
		log.Fatalf("Failed to create cluster '%s' because a cluster with that name already exists", clusterConfig.Cluster.Name)
	}

	// create cluster
	if clusterConfig.KubeconfigOpts.UpdateDefaultKubeconfig {
		log.Debugln("'--kubeconfig-update-default set: enabling wait-for-server")
		clusterConfig.ClusterCreateOpts.WaitForServer = true
	}

	if err := k3dCluster.ClusterRun(ctx, runtimes.SelectedRuntime, clusterConfig); err != nil {
		// rollback if creation failed
		log.Errorln(err)
		if cfg.Options.K3dOptions.NoRollback { // TODO: move rollback mechanics to pkg/
			log.Fatalln("Cluster creation FAILED, rollback deactivated.")
		}
		// rollback if creation failed
		log.Errorln("Failed to create cluster >>> Rolling Back")
		if err := k3dCluster.ClusterDelete(ctx, runtimes.SelectedRuntime, &clusterConfig.Cluster, k3d.ClusterDeleteOpts{SkipRegistryCheck: true}); err != nil {
			log.Errorln(err)
			log.Fatalln("Cluster creation FAILED, also FAILED to rollback changes!")
		}
		log.Fatalln("Cluster creation FAILED, all changes have been rolled back!")
	}
	log.Infof("Cluster '%s' created successfully!", clusterConfig.Cluster.Name)

	// get kubeconfig
	if clusterConfig.KubeconfigOpts.UpdateDefaultKubeconfig && clusterConfig.KubeconfigOpts.SwitchCurrentContext {
		log.Infoln("--kubeconfig-update-default=false --> sets --kubeconfig-switch-context=false")
		clusterConfig.KubeconfigOpts.SwitchCurrentContext = false
	}

	if clusterConfig.KubeconfigOpts.UpdateDefaultKubeconfig {
		log.Debugf("Updating default kubeconfig with a new context for cluster %s", clusterConfig.Cluster.Name)
		if _, err := k3dCluster.KubeconfigGetWrite(ctx, runtimes.SelectedRuntime, &clusterConfig.Cluster, "", &k3dCluster.WriteKubeConfigOptions{UpdateExisting: true, OverwriteExisting: false, UpdateCurrentContext: cfg.Options.KubeconfigOptions.SwitchCurrentContext}); err != nil {
			log.Warningln(err)
		}
	}

	// let user know what was done
	// print information on how to use the cluster with kubectl
	log.Infoln("You can now use it like this:")
	if clusterConfig.KubeconfigOpts.UpdateDefaultKubeconfig && !clusterConfig.KubeconfigOpts.SwitchCurrentContext {
		fmt.Printf("kubectl config use-context %s\n", fmt.Sprintf("%s-%s", k3d.DefaultObjectNamePrefix, clusterConfig.Cluster.Name))
	} else if !clusterConfig.KubeconfigOpts.SwitchCurrentContext {
		if runtime.GOOS == "windows" {
			fmt.Printf("$env:KUBECONFIG=(%s kubeconfig write %s)\n", os.Args[0], clusterConfig.Cluster.Name)
		} else {
			fmt.Printf("export KUBECONFIG=$(%s kubeconfig write %s)\n", os.Args[0], clusterConfig.Cluster.Name)
		}
	}
	fmt.Println("kubectl cluster-info")
}

// DeleteK3dCluster will delete a k3d cluster
func DeleteK3dCluster() {

	// k3dCluster.ClusterDelete()

}
