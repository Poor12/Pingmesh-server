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
	//nodelist map[string][]string
	//iptopatition map[string]string
	gstorage *GlobalStorage
)

func (c Config) Complete() (*PingmeshServer, error) {
	informer, err := c.informer()
	if err != nil {
		return nil, err
	}
	nodes := informer.Core().V1().Nodes()
	store := NewStorage()
	gstorage = NewGlobalStorage()
	scrape := NewScraper(nodes.Lister(), store)
	return &PingmeshServer{
		syncs:      []cache.InformerSynced{nodes.Informer().HasSynced},
		informer:   informer,
		scraper:    scrape,
		storage:    store,
		resolution: c.MetricResolution,
	}, nil
}

func (c Config) informer() (informers.SharedInformerFactory, error) {
	// set up the informers
	kubeClient, err := kubernetes.NewForConfig(c.Rest)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}
	// we should never need to resync, since we're not worried about missing events,
	// and resync is actually for regular interval-based reconciliation these days,
	// so set the default resync interval to 0
	return informers.NewSharedInformerFactory(kubeClient, 0), nil
}
