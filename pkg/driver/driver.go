package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/container-object-storage-interface/api/apis/cosi.sigs.k8s.io/v1alpha1"
	cs "github.com/container-object-storage-interface/api/client/clientset/typed/cosi.sigs.k8s.io/v1alpha1"
	"github.com/container-storage-interface/spec/lib/go/csi"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"os"
	"path/filepath"
)

// CosiDriver implements the NodePublishVolume and NodeUnpublishVolume methods
// of the csi.NodeServer interface
type CosiDriver struct {
	name       string
	nodeID     string
	cosiClient cs.CosiV1alpha1Interface
	ctx        context.Context
}

var _ csi.NodeServer = &CosiDriver{}

var _ csi.IdentityServer = &CosiDriver{}

func NewCosiDriver(nodeId, driverName string, c cs.CosiV1alpha1Interface) *CosiDriver {
	return &CosiDriver{
		name:       driverName,
		nodeID:     nodeId,
		cosiClient: c,
		ctx:        context.Background(),
	}
}

const protocolFileName string = `protocolConn.json`

func (d CosiDriver) NodePublishVolume(_ context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.Infof("NodePublishVolume: volId: %v, targetPath: %v\n", req.GetVolumeId(), req.GetTargetPath())

	n, ns, err := bucketAccessRequestNameNamespace(req.VolumeContext)
	if err != nil {
		return nil, err
	}

	getError := func(t, n string) error { return fmt.Errorf("failed to get %s: %s", t, n)}

	bar, err := d.cosiClient.BucketAccessRequests(ns).Get(d.ctx, n, v1.GetOptions{})
	if err != nil || bar == nil {
		return nil, getError("bucketAccessRequest", ns + `\` + n)
	}

	br, err := d.cosiClient.BucketRequests(bar.Namespace).Get(d.ctx, bar.Spec.BucketRequestName, v1.GetOptions{})
	if err != nil || br == nil {
		return nil, getError("bucketRequest", bar.Namespace + `\` + bar.Spec.BucketRequestName)
	}

	bkt, err := d.cosiClient.Buckets().Get(d.ctx, br.Spec.BucketName, v1.GetOptions{})
	if err != nil || bkt == nil{
		return nil, getError("bucket", br.Spec.BucketName)
	}

	var protocolConnection interface{}
	switch bkt.Spec.Protocol.ProtocolSignature {
	case v1alpha1.ProtocolSignatureS3:
		protocolConnection = bkt.Spec.Protocol.S3
	case v1alpha1.ProtocolSignatureAzure:
		protocolConnection = bkt.Spec.Protocol.Azure
	case v1alpha1.ProtocolSignatureGCS:
		protocolConnection = bkt.Spec.Protocol.GCS
	case "":
		err = fmt.Errorf("bucket protocol signature missing")
	default:
		err = fmt.Errorf("unrecognized protocol %q, unable to extract connection data", bkt.Spec.Protocol)
	}
	if err != nil {
		return nil, err
	}

	protoData, err := json.Marshal(protocolConnection)
	if err != nil {
		return nil, fmt.Errorf("error marshalling protocol: %v", err)
	}

	target := filepath.Join(req.TargetPath, protocolFileName)
	f, err := os.Open(target)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %s: %v", target, err)
	}
	defer f.Close()
	_, err = f.Write(protoData)
	if err != nil {
		return nil, fmt.Errorf("unable to write to file: %v", err)
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

const (
	barName      = "bucketAccessRequestName"
	barNamespace = "bucketAccessRequestNamespace"
)

func bucketAccessRequestNameNamespace(volCtx map[string]string) (name, ns string, err error) {

	e := func(m string) error { return fmt.Errorf("required volume context key unset: %v", m) }

	var ok bool
	name, ok = volCtx[barName]
	if ! ok {
		return "", "", e(barName)
	}
	ns, ok = volCtx[barNamespace]
	if ! ok {
		return "", "", e(barNamespace)
	}
	return
}

///////////////////////////////////
// Nodes Services               //
/////////////////////////////////

func (d CosiDriver) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.Infof("NodeUnpublishVolume: volId: %v, targetPath: %v\n", req.GetVolumeId(), req.GetTargetPath())
	target := filepath.Join(req.TargetPath, protocolFileName)
	err := os.RemoveAll(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to remove file %s: %v", target, err)
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (d CosiDriver) NodeGetInfo(_ context.Context, _ *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.Infof("NodeGetInfo()")
	resp := &csi.NodeGetInfoResponse{
		NodeId: d.nodeID,
	}
	return resp, nil

}

func (d CosiDriver) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {

	return nil, nil
}

func (d CosiDriver) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, nil
}

func (d CosiDriver) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, nil
}

func (d CosiDriver) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}

func (d CosiDriver) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, nil
}

///////////////////////////////////
// Identity Services            //
/////////////////////////////////

func (d CosiDriver) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.Infoln("GetPluginInfo")
	return &csi.GetPluginInfoResponse{
		Name:          d.name,
		VendorVersion: "v1alpha1",
	}, nil
}

func (d CosiDriver) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

func (d CosiDriver) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}
