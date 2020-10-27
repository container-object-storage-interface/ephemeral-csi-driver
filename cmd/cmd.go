/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/golang/glog"
)

var Version string

// flags
var (
	identity = "cosi.storage.k8s.io"
	nodeID   = ""
	protocol = ""
	listen   = ""
	//endpoint = "unix://csi/csi.sock"
)

var driverCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Ephemeral CSI driver for use in the COSI",
	Long: "This Container Storage Interface (CSI) driver provides the ability to reference Bucket and BucketAccess objects, extracting connection/credential information and writing it to the Pod's filesystem. This driver does not manage the lifecycle of the bucket or the backing of the objects themselves, it only acts as the middle-man.",
	SilenceUsage: true,
	RunE: func(c *cobra.Command, args []string) error {
		return driver(args)
	},
}

func init() {
	if Version == "" {
		Version = "dev"
	}

	viper.AutomaticEnv()
	// parse the go default flagset to get flags for glog and other packages in future
	driverCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// defaulting this to true so that logs are printed to console
	flag.Set("logtostderr", "true")

	driverCmd.PersistentFlags().StringVarP(&identity, "identity", "i", identity, "identity of this COSI CSI driver")
	//driverCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", endpoint, "endpoint at which COSI CSI driver is listening")
	driverCmd.PersistentFlags().StringVarP(&nodeID, "node-id", "n", nodeID, "identity of the node in which COSI CSI driver is running")
	driverCmd.PersistentFlags().StringVarP(&listen, "listen", "l", listen, "address of the listening socket for the node server")
	driverCmd.PersistentFlags().StringVarP(&protocol, "protocol", "p", protocol, "must be one of tcp, tcp4, tcp6, unix, unixpacket")

	driverCmd.PersistentFlags().MarkHidden("alsologtostderr")
	driverCmd.PersistentFlags().MarkHidden("log_backtrace_at")
	driverCmd.PersistentFlags().MarkHidden("log_dir")
	driverCmd.PersistentFlags().MarkHidden("logtostderr")
	driverCmd.PersistentFlags().MarkHidden("master")
	driverCmd.PersistentFlags().MarkHidden("stderrthreshold")
	driverCmd.PersistentFlags().MarkHidden("vmodule")

	// suppress the incorrect prefix in glog output
	flag.CommandLine.Parse([]string{})
	viper.BindPFlags(driverCmd.PersistentFlags())
}

func Execute() error {
	return driverCmd.Execute()
}
