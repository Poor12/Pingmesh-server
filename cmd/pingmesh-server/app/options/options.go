package options

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	pingmesh_server "pingmesh-server/pkg/pingmesh-server"
	"time"
)

type Options struct {
	Kubeconfig       string
	MetricResolution time.Duration
}

func NewOptions() *Options {
	o := &Options{
		MetricResolution: 60 * time.Second,
	}

	return o
}

func (o Options) PingmeshServerConfig() (*pingmesh_server.Config, error) {
	restConfig, err := o.restConfig()
	if err != nil {
		return nil, err
	}
	return &pingmesh_server.Config{
		Rest:             restConfig,
		MetricResolution: o.MetricResolution,
	}, nil
}

func (o Options) restConfig() (*rest.Config, error) {
	var clientConfig *rest.Config
	var err error
	if len(o.Kubeconfig) > 0 {
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: o.Kubeconfig}
		loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

		clientConfig, err = loader.ClientConfig()
	} else {
		clientConfig, err = rest.InClusterConfig()
		//kubeconfig := filepath.Join("/Users/tiechengshen/", ".kube", "config")
		//clientConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client config: %v", err)
	}
	return clientConfig, err
}
