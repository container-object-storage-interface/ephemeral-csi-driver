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
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"

	"github.com/container-object-storage-interface/ephemeral-csi-driver/pkg/controller"
	id "github.com/container-object-storage-interface/ephemeral-csi-driver/pkg/identity"

	"github.com/golang/glog"
)

func driver(args []string) error {
	idServer, err := id.NewIdentityServer(identity, Version, map[string]string{})
	if err != nil {
		return err
	}
	glog.V(5).Infof("identity server started")

	ctrlServer, err := controller.NewControllerServer(identity, nodeID)
	if err != nil {
		return err
	}
	glog.V(5).Infof("controller manager started")

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(endpoint, idServer, ctrlServer, nil)
	s.Wait()

	return nil
}