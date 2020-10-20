package pingmesh_server

import (
	"context"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type PingmeshServer struct {
	syncs    []cache.InformerSynced
	informer informers.SharedInformerFactory

	resolution    time.Duration
	scraper       *Scraper
	storage       *Storage
	healthMu      sync.RWMutex
	lastTickStart time.Time
	lastOk        bool
}

func (pm *PingmeshServer) RunUntil(stopCh <-chan struct{}) error {
	pm.informer.Start(stopCh)
	shutdown := cache.WaitForCacheSync(stopCh, pm.syncs...)
	if !shutdown {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startHTTPServer()
	return pm.runScrape(ctx)
}

func (pm *PingmeshServer) runScrape(ctx context.Context) error {
	ticker := time.NewTicker(pm.resolution)
	defer ticker.Stop()
	pm.scrape(ctx, time.Now())

	for {
		select {
		case startTime := <-ticker.C:
			pm.scrape(ctx, startTime)
		case <-ctx.Done():
			return nil
		}
	}
}

func (pm *PingmeshServer) scrape(ctx context.Context, startTime time.Time) {
	pm.healthMu.Lock()
	pm.lastTickStart = startTime
	pm.healthMu.Unlock()

	healthyTick := true

	ctx, cancelTimeout := context.WithTimeout(ctx, pm.resolution)
	defer cancelTimeout()

	klog.V(6).Infof("Beginning cycle, checking nodes...")
	//storage = NewStorage()

	err := pm.scraper.Scrape(ctx)
	if err != nil {
		healthyTick = false
	}

	pm.healthMu.Lock()
	pm.lastOk = healthyTick
	pm.healthMu.Unlock()
}
