package pingmesh_server

import "sync"

type Storage struct {
	mu            sync.RWMutex
	nodesPatition map[string][]string
	IPtoPatition  map[string]string
	//metricsData   map[string][]*ProberResultOne
	//pinglist      map[string][]string
}

type GlobalStorage struct {
	mu            sync.RWMutex
	nodesPatition map[string][]string
	IPtoPatition  map[string]string
	metricsData   map[string][]*ProberResultOne
}
func NewStorage() *Storage {
	return &Storage{
		nodesPatition: make(map[string][]string),
		IPtoPatition: make(map[string]string),
		//metricsData: make(map[string][]*ProberResultOne),
	}
}

func NewGlobalStorage() *GlobalStorage {
	return &GlobalStorage{
		nodesPatition: make(map[string][]string),
		IPtoPatition: make(map[string]string),
		metricsData: make(map[string][]*ProberResultOne),
	}
}