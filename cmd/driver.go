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
	"net"
	"os"

	cs "github.com/container-object-storage-interface/api/clientset/typed/objectstorage.k8s.io/v1alpha1"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	id "github.com/container-object-storage-interface/ephemeral-csi-driver/pkg/identity"
	"github.com/container-object-storage-interface/ephemeral-csi-driver/pkg/node"
)

func driver(args []string) error {

	if protocol == "unix" {
		if err := os.RemoveAll(listen); err != nil {
			klog.Fatalf("could not prepare socket: %v", err)
		}
	}

	idServer, err := id.NewIdentityServer(identity, Version, map[string]string{})
	if err != nil {
		return err
	}
	glog.V(5).Infof("identity server started")

	config := &rest.Config{}

	client := cs.NewForConfigOrDie(config)

	node.Initalize(basePath)
	node := node.NewNodeServer(identity, nodeID, *client)
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	csi.RegisterNodeServer(srv, node)
	csi.RegisterIdentityServer(srv, idServer)
	l, err := net.Listen(protocol, listen)
	if err != nil {
		klog.Fatalf("could not create listener: %v", err)
	}
	if err = srv.Serve(l); err != nil {
		klog.Fatalf("%v", err)
	}

	return nil
}