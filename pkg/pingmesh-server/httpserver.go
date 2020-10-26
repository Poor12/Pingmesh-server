package pingmesh_server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

func startHTTPServer() {
	router := mux.NewRouter()
	router.HandleFunc(DefaultPingmeshDownloadURL, pinglistHandler).Methods("Get")
	router.HandleFunc(DefaultPingmeshUploadURL,metricsHandler).Methods("POST")
	addr := fmt.Sprintf("%s:%d", DefaultHttpsAddress, DefaultHttpsPort)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	klog.Info("testing.........")
	klog.Fatal(server.ListenAndServe())
}

type pinglist struct {
	WorkerName           string
	Patition			 string
	PingList             map[string][]string
}

type ProberResultOne struct {
	WorkerName           string
	MetricName           string
	TargetAddr           string
	SourceRegion         string
	TargetRegion         string
	ProbeType            string
	TimeStamp            int64
	Value                float32
}

func pinglistHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r)

	pl := &pinglist{
		WorkerName: strings.Split(r.RemoteAddr,":")[0],
		Patition: gstorage.IPtoPatition[strings.Split(r.RemoteAddr,":")[0]],
		PingList: gstorage.nodesPatition,
	}

	pljson, _ := json.Marshal(pl)
	w.Write([]byte(pljson))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {

	//pros := make([]([]*ProberResultOne),0)
	metrics := &ProberResultOne{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil{
		w.Write([]byte("fail"))
		return
	}
	err = json.Unmarshal(data, metrics)
	if err != nil{
		w.Write([]byte("fail"))
		return
	}

	gstorage.mu.Lock()
	wn := &gstorage.metricsData
	if wn == nil{
		(*wn)[metrics.WorkerName] = make([]*ProberResultOne,0)
	}
	(*wn)[metrics.WorkerName] = append((*wn)[metrics.WorkerName],metrics)
	gstorage.mu.Unlock()
	w.Write([]byte("success"))
	//w.Write([]byte("127.0.0.2"))
}