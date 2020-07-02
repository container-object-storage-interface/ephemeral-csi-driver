package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"net"
	"os"

	"cosi-csi-driver/pkg/driver"

	cs "github.com/container-object-storage-interface/api/client/clientset/typed/cosi.sigs.k8s.io/v1alpha1"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog"
)

var (
	nodeId     string
	listen     string
	protocol   string
	master     string
	kubeconfig string
)

const driverName = "csi-cosi-adapter"

const (
	nodeIdOpt     = "nodeid"
	listenOpt     = "listen"
	protocolOpt   = "protocol"
	masterURLOpt  = "master"
	kubeconfigOpt = "kubeconfig"
)

var requiredFlags = map[string]struct{}{
	nodeIdOpt:   {},
	listenOpt:   {},
	protocolOpt: {},
}

var fs *flag.FlagSet

func checkRequiredFlags(fs *flag.FlagSet) error {
	unsetRequiredFlags := make([]string, 0)

	fs.VisitAll(func(f *flag.Flag) {
		klog.Infof("flag %v: %v\n", f.Name, f.Value)
		if _, ok := requiredFlags[f.Name]; ok {
			if len(f.Value.String()) == 0 {
				unsetRequiredFlags = append(unsetRequiredFlags, f.Name)
			}
		}
	})
	if len(unsetRequiredFlags) > 0 {
		return fmt.Errorf("following unset flags are required: %v", unsetRequiredFlags)
	}
	return nil
}

func init() {
	fs = flag.NewFlagSet("cosi", flag.PanicOnError)

	fs.StringVar(&kubeconfig, kubeconfigOpt, "", "path to kubeconfig")
	fs.StringVar(&listen, listenOpt, "", "listening socket of this server")
	fs.StringVar(&nodeId, nodeIdOpt, "", "pod.metadata.node")
	fs.StringVar(&master, masterURLOpt, "", "url of API server")
	fs.StringVar(&protocol, protocolOpt, "", "Must be one of tcp, tcp4, tcp6, unix, unixpacket")

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
	if err := checkRequiredFlags(fs); err != nil {
		panic(err)
	}
}

func main() {
	defer klog.Flush()

	if protocol == "unix" {
		if err := os.RemoveAll(listen); err != nil {
			klog.Fatalf("could not prepare socket: %v", err)
		}
	}

	cfg, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	cc := cs.NewForConfigOrDie(cfg)

	d := driver.NewCosiDriver(nodeId, driverName, cc)

	srv := grpc.NewServer()
	csi.RegisterNodeServer(srv, d)
	csi.RegisterIdentityServer(srv, d)
	l, err := net.Listen(protocol, listen)
	if err != nil {
		klog.Fatalf("could not create listener: %v", err)
	}
	if err = srv.Serve(l); err != nil {
		klog.Fatalf("%v", err)
	}
}
