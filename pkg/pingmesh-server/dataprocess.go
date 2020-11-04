package pingmesh_server

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

const (
	MetricCollectInterval      = 60 * time.Second
	TargetFlushManagerInterval = 60 * time.Second
	MetricOriginSeparator      = `_`
	MetricUniqueSeparator      = `#`
)

var (
	//IcmpDataMap = sync.Map{}
	//HttpDataMap         = sync.Map{}
	PingLatencyGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsNamePingLatency,
		Help: "Duration of ping prober ",
	}, []string{"source_region", "target_region"})
	PingPackageDropGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsNamePingPackageDrop,
		Help: "rate of ping packagedrop ",
	}, []string{"source_region", "target_region"})

	PingTargetSuccessGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsNamePingTargetSuccess,
		Help: "target success",
	}, []string{"source_region", "target_region"})
)

func (pm *PingmeshServer) NewMetrics() {
	prometheus.DefaultRegisterer.MustRegister(PingLatencyGaugeVec)
	prometheus.DefaultRegisterer.MustRegister(PingPackageDropGaugeVec)
	prometheus.DefaultRegisterer.MustRegister(PingTargetSuccessGaugeVec)
}

func (pm *PingmeshServer) DataProcess(ctx context.Context) error {
	ticker := time.NewTicker(pm.resolution)
	klog.Info("msg: DataProcessManager start....")
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:

			go pm.IcmpDataProcess()
			//go HttpDataProcess()

		case <-ctx.Done():
			klog.Info("msg: DataProcessManager exit....")
			return nil
		}
	}

	return nil
}

func (pm *PingmeshServer) IcmpDataProcess() {

	klog.Info("msg: IcmpDataProcess run....")

	var expireds []string

	latencyMap := make(map[string][]float64)
	packagedropMap := make(map[string][]float64)
	targetSuccMap := make(map[string][]float64)

	f := func(k, v interface{}) bool {
		key := k.(string)
		va := v.(*ProberResultOne)

		// check item expire
		now := time.Now().Unix()
		if now-va.TimeStamp > 300 {
			expireds = append(expireds, key)
		} else {
			if strings.Contains(va.MetricName, MetricOriginSeparator) {
				metricType := strings.Split(va.MetricName, MetricOriginSeparator)[1]
				uniqueKey := va.MetricName + MetricUniqueSeparator + va.SourceRegion + MetricUniqueSeparator + va.TargetRegion

				switch metricType {
				case "latency":
					old := latencyMap[uniqueKey]
					if len(old) == 0 {
						latencyMap[uniqueKey] = []float64{float64(va.Value)}
					} else {
						latencyMap[uniqueKey] = append(latencyMap[uniqueKey], float64(va.Value))
					}
				case "packageDrop":
					old := packagedropMap[uniqueKey]
					if len(old) == 0 {
						packagedropMap[uniqueKey] = []float64{float64(va.Value)}
					} else {
						packagedropMap[uniqueKey] = append(packagedropMap[uniqueKey], float64(va.Value))
					}
				case "target":
					old := targetSuccMap[uniqueKey]
					if len(old) == 0 {
						targetSuccMap[uniqueKey] = []float64{float64(va.Value)}
					} else {
						targetSuccMap[uniqueKey] = append(targetSuccMap[uniqueKey], float64(va.Value))
					}
				}
			}

		}

		return true
	}

	pm.storage.mu.Lock()
	pm.storage.metricsData.Range(f)
	// delete  expireds
	for _, e := range expireds {
		pm.storage.metricsData.Delete(e)
	}
	pm.storage.mu.Unlock()

	// compute data with avg or pct99
	dealWithDataMapAvg(latencyMap, PingLatencyGaugeVec, "icmp")
	dealWithDataMapAvg(packagedropMap, PingPackageDropGaugeVec, "icmp")

	dealWithDataMapBool(targetSuccMap, PingTargetSuccessGaugeVec, "icmp")

}

func dealWithDataMapAvg(dataM map[string][]float64, promeVec *prometheus.GaugeVec, pType string) {
	for uniqueKey, datas := range dataM {
		//MetricName := strings.Split(uniqueKey, MetricUniqueSeparator)[0]
		SourceRegion := strings.Split(uniqueKey, MetricUniqueSeparator)[1]
		TargetRegionOrAddr := strings.Split(uniqueKey, MetricUniqueSeparator)[2]
		var sum, avg float64
		num := len(datas)
		for _, ds := range datas {
			sum += ds
		}
		avg = sum / float64(num)
		//klog.Infof("AvgPingLatency from %s to %s: %f",SourceRegion,TargetRegionOrAddr,avg)
		switch pType {
		case "http":
			promeVec.With(prometheus.Labels{"source_region": SourceRegion, "addr": TargetRegionOrAddr}).Set(avg)
		case "icmp":
			promeVec.With(prometheus.Labels{"source_region": SourceRegion, "target_region": TargetRegionOrAddr}).Set(avg)
		}

	}
}

func dealWithDataMapBool(dataM map[string][]float64, promeVec *prometheus.GaugeVec, pType string) {

	for uniqueKey, datas := range dataM {
		//MetricName := strings.Split(uniqueKey, MetricUniqueSeparator)[0]
		SourceRegion := strings.Split(uniqueKey, MetricUniqueSeparator)[1]
		TargetRegionOrAddr := strings.Split(uniqueKey, MetricUniqueSeparator)[2]
		//var sum, avg float64
		//num := len(datas)

		thisFailNum := 0

		for _, ds := range datas {
			if ds == -1 {
				thisFailNum += 1
			}
		}

		if thisFailNum == len(datas) {
			promeVec.With(prometheus.Labels{"source_region": SourceRegion, "target_region": TargetRegionOrAddr}).Set(0)
			klog.Infof("the network from %s to %s has been disconnected",SourceRegion,TargetRegionOrAddr)
		} else {
			promeVec.With(prometheus.Labels{"source_region": SourceRegion, "target_region": TargetRegionOrAddr}).Set(1)
			klog.Infof("the network from %s to %s has been connected",SourceRegion,TargetRegionOrAddr)
		}

	}
}
