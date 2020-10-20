package main

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/component-base/logs"
	"os"
	"pingmesh-server/cmd/pingmesh-server/app"
	"runtime"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	cmd := app.NewPingmeshServerCommand(wait.NeverStop)
	//cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
