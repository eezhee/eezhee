package k3d

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/docker/go-connections/nat"
	cliutil "github.com/rancher/k3d/v4/cmd/util"
	k3dCluster "github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/config"
	conf "github.com/rancher/k3d/v4/pkg/config/v1alpha2"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	k3d "github.com/rancher/k3d/v4/pkg/types"
	log "github.com/sirupsen/logrus"
)

// CreateK3dCluster creates a local cluster
func CreateK3dCluster() {

	// TODO - temp while testing
	log.SetLevel(log.DebugLevel)

	// create a config
	var cfg conf.SimpleConfig

	cfg.Servers = 1
	cfg.Agents = 0
	cfg.Name = k3d.DefaultClusterName
	// cfg.Name = "k3d-eezhee"
	cfg.Image = k3d.DefaultK3sImageRepo + ":latest"
	// cfg.Image = "docker.io/rancher/k3s:latest" //TODO make try and get a list like we do with k3s binaries
	cfg.Registries.Create = false
	// cfg.Options.K3dOptions.DisableImageVolume = false

	cfg.Options.KubeconfigOptions.UpdateDefaultKubeconfig = true
	cfg.Options.KubeconfigOptions.SwitchCurrentContext = true

	// cfg
	// var (
	// 	exposeAPI *k3d.ExposureOpts
	// 	err       error
	// )

	// exposeAPI = &k3d.ExposureOpts{
	// 	PortMapping: nat.PortMapping{
	// 		Binding: nat.PortBinding{
	// 			HostIP:   cfg.ExposeAPI.HostIP,
	// 			HostPort: cfg.ExposeAPI.HostPort,
	// 		},
	// 	},
	// 	Host: cfg.ExposeAPI.Host,
	// }

	// if len(exposeAPI.Binding.HostPort) == 0 {
	// 	exposeAPI, err = cliutil.ParsePortExposureSpec("random", k3d.DefaultAPIPort)
	// 	if err != nil {
	// 		return
	// 	}
	// }

	// cfg.ExposeAPI = conf.SimpleExposureOpts{
	// 	Host:     exposeAPI.Host,
	// 	HostIP:   exposeAPI.Binding.HostIP,
	// 	HostPort: exposeAPI.Binding.HostPort,
	// }

	// servers - # of servers to create
	// agents - # of agents to create
	// image - which image to use for nodes
	// network - join existing network
	// token - cluster token
	// and lots more in k3d/cmd/cluster/clusterCreate.go

	cfg, err := applyCLIOverrides(cfg)
	if err != nil {
		log.Fatalf("Failed to apply CLI overrides: %+v", err)
	}

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

func applyCLIOverrides(cfg conf.SimpleConfig) (conf.SimpleConfig, error) {

	/****************************
	 * Parse and validate flags *
	 ****************************/

	// -> API-PORT
	// parse the port mapping
	var (
		err       error
		exposeAPI *k3d.ExposureOpts
	)

	// Apply config file values as defaults
	exposeAPI = &k3d.ExposureOpts{
		PortMapping: nat.PortMapping{
			Binding: nat.PortBinding{
				HostIP:   cfg.ExposeAPI.HostIP,
				HostPort: cfg.ExposeAPI.HostPort,
			},
		},
		Host: cfg.ExposeAPI.Host,
	}

	// Overwrite if cli arg is set
	// if ppViper.IsSet("cli.api-port") {
	// 	if cfg.ExposeAPI.HostPort != "" {
	// 		log.Debugf("Overriding pre-defined kubeAPI Exposure Spec %+v with CLI argument %s", cfg.ExposeAPI, ppViper.GetString("cli.api-port"))
	// 	}
	// 	exposeAPI, err = cliutil.ParsePortExposureSpec(ppViper.GetString("cli.api-port"), k3d.DefaultAPIPort)
	// 	if err != nil {
	// 		return cfg, err
	// 	}
	// }

	// Set to random port if port is empty string
	if len(exposeAPI.Binding.HostPort) == 0 {
		exposeAPI, err = cliutil.ParsePortExposureSpec("random", k3d.DefaultAPIPort)
		if err != nil {
			return cfg, err
		}
	}

	cfg.ExposeAPI = conf.SimpleExposureOpts{
		Host:     exposeAPI.Host,
		HostIP:   exposeAPI.Binding.HostIP,
		HostPort: exposeAPI.Binding.HostPort,
	}

	// -> VOLUMES
	// volumeFilterMap will map volume mounts to applied node filters
	// volumeFilterMap := make(map[string][]string, 1)
	// for _, volumeFlag := range ppViper.GetStringSlice("cli.volumes") {

	// 	// split node filter from the specified volume
	// 	volume, filters, err := cliutil.SplitFiltersFromFlag(volumeFlag)
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}

	// 	if strings.Contains(volume, k3d.DefaultRegistriesFilePath) && (cfg.Registries.Create || cfg.Registries.Config != "" || len(cfg.Registries.Use) != 0) {
	// 		log.Warnf("Seems like you're mounting a file at '%s' while also using a referenced registries config or k3d-managed registries: Your mounted file will probably be overwritten!", k3d.DefaultRegistriesFilePath)
	// 	}

	// 	// create new entry or append filter to existing entry
	// 	if _, exists := volumeFilterMap[volume]; exists {
	// 		volumeFilterMap[volume] = append(volumeFilterMap[volume], filters...)
	// 	} else {
	// 		volumeFilterMap[volume] = filters
	// 	}
	// }

	// for volume, nodeFilters := range volumeFilterMap {
	// 	cfg.Volumes = append(cfg.Volumes, conf.VolumeWithNodeFilters{
	// 		Volume:      volume,
	// 		NodeFilters: nodeFilters,
	// 	})
	// }

	// log.Tracef("VolumeFilterMap: %+v", volumeFilterMap)

	// -> PORTS
	// portFilterMap := make(map[string][]string, 1)
	// for _, portFlag := range ppViper.GetStringSlice("cli.ports") {
	// 	// split node filter from the specified volume
	// 	portmap, filters, err := cliutil.SplitFiltersFromFlag(portFlag)
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}

	// 	if len(filters) > 1 {
	// 		log.Fatalln("Can only apply a Portmap to one node")
	// 	}

	// 	// create new entry or append filter to existing entry
	// 	if _, exists := portFilterMap[portmap]; exists {
	// 		log.Fatalln("Same Portmapping can not be used for multiple nodes")
	// 	} else {
	// 		portFilterMap[portmap] = filters
	// 	}
	// }

	// for port, nodeFilters := range portFilterMap {
	// 	cfg.Ports = append(cfg.Ports, conf.PortWithNodeFilters{
	// 		Port:        port,
	// 		NodeFilters: nodeFilters,
	// 	})
	// }

	// log.Tracef("PortFilterMap: %+v", portFilterMap)

	// --label
	// labelFilterMap will add container label to applied node filters
	// labelFilterMap := make(map[string][]string, 1)
	// for _, labelFlag := range ppViper.GetStringSlice("cli.labels") {

	// 	// split node filter from the specified label
	// 	label, nodeFilters, err := cliutil.SplitFiltersFromFlag(labelFlag)
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}

	// 	// create new entry or append filter to existing entry
	// 	if _, exists := labelFilterMap[label]; exists {
	// 		labelFilterMap[label] = append(labelFilterMap[label], nodeFilters...)
	// 	} else {
	// 		labelFilterMap[label] = nodeFilters
	// 	}
	// }

	// for label, nodeFilters := range labelFilterMap {
	// 	cfg.Labels = append(cfg.Labels, conf.LabelWithNodeFilters{
	// 		Label:       label,
	// 		NodeFilters: nodeFilters,
	// 	})
	// }

	// log.Tracef("LabelFilterMap: %+v", labelFilterMap)

	// --env
	// envFilterMap will add container env vars to applied node filters
	// envFilterMap := make(map[string][]string, 1)
	// for _, envFlag := range ppViper.GetStringSlice("cli.env") {

	// 	// split node filter from the specified env var
	// 	env, filters, err := cliutil.SplitFiltersFromFlag(envFlag)
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}

	// 	// create new entry or append filter to existing entry
	// 	if _, exists := envFilterMap[env]; exists {
	// 		envFilterMap[env] = append(envFilterMap[env], filters...)
	// 	} else {
	// 		envFilterMap[env] = filters
	// 	}
	// }

	// for envVar, nodeFilters := range envFilterMap {
	// 	cfg.Env = append(cfg.Env, conf.EnvVarWithNodeFilters{
	// 		EnvVar:      envVar,
	// 		NodeFilters: nodeFilters,
	// 	})
	// }

	// log.Tracef("EnvFilterMap: %+v", envFilterMap)

	return cfg, nil
}
