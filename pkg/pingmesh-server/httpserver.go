package pingmesh_server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

var (
	kubeClient *kubernetes.Clientset
	gstorage   *Storage
)

type pinglist struct {
	WorkerName string
	Patition   string
	PingList   map[string][]string
}

type ProberResultOne struct {
	WorkerName   string
	MetricName   string
	TargetAddr   string
	SourceRegion string
	TargetRegion string
	ProbeType    string
	TimeStamp    int64
	Value        float32
}

func (pm *PingmeshServer) startHTTPServer() {
	router := mux.NewRouter()
	kubeClient = pm.kubeClient
	gstorage = pm.storage
	router.HandleFunc(DefaultPingmeshDownloadURL, pinglistHandler).Methods("Get")
	router.HandleFunc(DefaultPingmeshUploadURL, metricsHandler).Methods("POST")
	addr := fmt.Sprintf("%s:%d", DefaultHttpsAddress, DefaultHttpsPort)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	klog.Fatal(server.ListenAndServe())
}

func (pm *PingmeshServer) startMetrics() {
	pm.NewMetrics()
	http.Handle("/metrics", promhttp.Handler())
	webListenAddr := fmt.Sprintf("%s:%d", DefaultHttpsAddress, DefaultMetricsPort)
	srv := http.Server{Addr: webListenAddr}
	//klog.Info("msg: Listening on address, address: ", webListenAddr)
	klog.Fatal(srv.ListenAndServe())
}

func pinglistHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r)

	podIP := strings.Split(r.RemoteAddr, ":")[0]
	selector := fields.OneTermEqualSelector("status.podIP", podIP).String()
	//set := fields.SelectorFromSet(selector)
	pods, err := kubeClient.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{FieldSelector: selector})
	if err != nil{
		klog.Error("cannot resolve podIP.....")
		return
	}
	pl := &pinglist{
		WorkerName: pods.Items[0].Status.HostIP,
		Patition:   gstorage.IPtoPatition[pods.Items[0].Status.HostIP],
		PingList:   gstorage.nodesPatition,
	}

	pljson, _ := json.Marshal(pl)
	w.Write([]byte(pljson))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {

	//pros := make([]([]*ProberResultOne),0)
	metrics := &ProberResultOne{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("fail"))
		return
	}
	err = json.Unmarshal(data, metrics)
	if err != nil {
		w.Write([]byte("fail"))
		return
	}

	gstorage.PushProberResults(metrics)

	//klog.Info("Success: ", gstorage.metricsData)
	w.Write([]byte("success"))
	//w.Write([]byte("127.0.0.2"))
}
