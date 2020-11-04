package pingmesh_server

import (
	"fmt"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

type Config struct {
	Rest             *rest.Config
	MetricResolution time.Duration
}

var (
//gstorage *GlobalStorage
)

func (c Config) Complete() (*PingmeshServer, error) {
	kubeClient, err := c.client()
	if err != nil {
		return nil, err
	}
	informer, err := c.informer(kubeClient)
	if err != nil {
		return nil, err
	}
	nodes := informer.Core().V1().Nodes()
	store := NewStorage()
	//gstorage = NewGlobalStorage()
	scrape := NewScraper(nodes.Lister(), store)
	return &PingmeshServer{
		syncs:      []cache.InformerSynced{nodes.Informer().HasSynced},
		kubeClient: kubeClient,
		informer:   informer,
		scraper:    scrape,
		storage:    store,
		resolution: c.MetricResolution,
	}, nil
}

func (c Config) client() (*kubernetes.Clientset, error) {
	kubeClient, err := kubernetes.NewForConfig(c.Rest)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}
	return kubeClient, nil
}

func (c Config) informer(kubeClient *kubernetes.Clientset) (informers.SharedInformerFactory, error) {
	// we should never need to resync, since we're not worried about missing events,
	// and resync is actually for regular interval-based reconciliation these days,
	// so set the default resync interval to 0
	//kubeClient, _ :=c.client()
	return informers.NewSharedInformerFactory(kubeClient, 0), nil
}
