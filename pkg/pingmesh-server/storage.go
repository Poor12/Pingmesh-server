package pingmesh_server

import (
	"sync"
)

type Storage struct {
	mu            sync.RWMutex
	nodesPatition map[string][]string
	IPtoPatition  map[string]string
	//metricsData   map[string][]*ProberResultOne
	metricsData sync.Map
	//pinglist      map[string][]string

}

//type GlobalStorage struct {
//	mu            sync.RWMutex
//	nodesPatition map[string][]string
//	IPtoPatition  map[string]string
//	//metricsData   map[string][]*ProberResultOne
//	metricsData sync.Map
//}

func NewStorage() *Storage {
	return &Storage{
		nodesPatition: make(map[string][]string),
		IPtoPatition:  make(map[string]string),
		//metricsData: make(map[string][]*ProberResultOne),
	}
}

//func NewGlobalStorage() *GlobalStorage {
//	return &GlobalStorage{
//		nodesPatition: make(map[string][]string),
//		IPtoPatition:  make(map[string]string),
//		//metricsData: make(map[string][]*ProberResultOne),
//		metricsData: sync.Map{},
//	}
//}

func GetProbeResultUid(prr *ProberResultOne) (uid string) {
	uid = prr.WorkerName + "-" + prr.MetricName + "-" + prr.SourceRegion + "-" + prr.TargetRegion + "-" + prr.ProbeType + "-" + prr.TargetAddr
	return
}

func (s *Storage) PushProberResults(in *ProberResultOne) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	suNum := 0
	//for _, prr := range in{
	uid := GetProbeResultUid(in)
	switch in.ProbeType {
	case `icmp`:
		s.metricsData.Store(uid, in)
		//case `http`:
		//	HttpDataMap.Store(uid, prr)
	}
	suNum += 1
	//}
	return nil
}
