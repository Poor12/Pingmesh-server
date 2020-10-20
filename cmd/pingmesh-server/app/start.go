package app

import (
	"github.com/spf13/cobra"
	"pingmesh-server/cmd/pingmesh-server/app/options"
)

func NewPingmeshServerCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch pingmesh-server",
		Long:  "Launch pingmesh-server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := runCommand(opts, stopCh); err != nil {
				return err
			}
			return nil
		},
	}
	//opts.Flags(cmd)
	return cmd
}

func runCommand(o *options.Options, ch <-chan struct{}) error {
	//if o.ShowVersion {
	//	fmt.Println(version.VersionInfo())
	//	os.Exit(0)
	//}
	config, err := o.PingmeshServerConfig()
	if err != nil {
		return err
	}
	// Use protobufs for communication with apiserver
	config.Rest.ContentType = "application/vnd.kubernetes.protobuf"

	pm, err := config.Complete()
	if err != nil {
		return err
	}

	//err = ms.AddHealthChecks(healthz.NamedCheck("healthz", ms.CheckHealth))
	//if err != nil {
	//	return err
	//}
	return pm.RunUntil(ch)
}
