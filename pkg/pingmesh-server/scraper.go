package pingmesh_server

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
	"sync"
)

func NewScraper(nodeLister v1listers.NodeLister, store *Storage) *Scraper {
	return &Scraper{
		nodeLister:   nodeLister,
		storage:      store,
		addrResolver: NewPriorityNodeAddressResolver(DefaultAddressTypePriority),
	}
}

type Scraper struct {
	nodeLister   v1listers.NodeLister
	storage      *Storage
	addrResolver NodeAddressResolver
}

var waitGroup sync.WaitGroup

func (c *Scraper) Scrape(baseCtx context.Context) error {
	set := labels.Set{NodeRoleKey: NodeRoleValue}
	selector := labels.SelectorFromSet(set)
	nodes, err := c.nodeLister.List(selector)
	var errs []error
	errChannel := make(chan error, len(nodes))
	if err != nil {
		// save the error, and continue on in case of partial results
		errs = append(errs, err)
	}
	klog.V(1).Infof("Scraping metrics from %v nodes", len(nodes))

	c.storage.mu.Lock()
	c.storage.nodesPatition = make(map[string][]string, 0)
	c.storage.mu.Unlock()

	for _, node := range nodes {
		waitGroup.Add(1)
		go func(node *corev1.Node) {
			// Prevents network congestion.
			addr, err := c.addrResolver.NodeAddress(node)
			if err != nil {
				err = fmt.Errorf("unable to extract connection information for node %q: %v", node.Name, err)
				errChannel <- err
				return
			}
			//fmt.Println(c.addrResolver.NodeAddress(node))
			p := node.ObjectMeta.Labels["patition"]
			if p == "" {
				err = fmt.Errorf("Node %q has not been partitioned", node.Name)
				errChannel <- err
				return
			}

			c.storage.mu.Lock()
			defer c.storage.mu.Unlock()

			ps := &c.storage.nodesPatition
			iptop := &c.storage.IPtoPatition
			if (*ps)[p] == nil {
				(*ps)[p] = []string{}
			}
			(*ps)[p] = append((*ps)[p], addr)
			(*iptop)[addr] = p
			errChannel <- err
			waitGroup.Done()
		}(node)
	}

	waitGroup.Wait()

	//c.storage.mu.Lock()
	//c.storage.nodesPatition = c.storage.nodesPatition
	//gstorage.IPtoPatition = c.storage.IPtoPatition
	//gstorage.mu.Unlock()

	for range nodes {
		err := <-errChannel
		if err != nil {
			errs = append(errs, err)
			// NB: partial node results are still worth saving, so
			// don't skip storing results if we got an error
		}
	}

	return utilerrors.NewAggregate(errs)
}
