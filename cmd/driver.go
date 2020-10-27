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
	"github.com/container-object-storage-interface/ephemeral-csi-driver/pkg/controller"
	"os"

	cs "github.com/container-object-storage-interface/api/clientset/typed/objectstorage.k8s.io/v1alpha1"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/client-go/kubernetes"
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
	kube := kubernetes.NewForConfigOrDie(config)

	nodeServer := node.NewNodeServer(identity, nodeID, *client, kube)
	controllerServer, err := controller.NewControllerServer(identity, nodeID)

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(listen, idServer, controllerServer, nodeServer)
	s.Wait()

	return nil
}