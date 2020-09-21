package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/container-object-storage-interface/api/apis/storage.k8s.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"os"
	"path/filepath"

	cs "github.com/container-object-storage-interface/api/clientset/typed/storage.k8s.io/v1alpha1"
	"github.com/container-storage-interface/spec/lib/go/csi"
)

var _ csi.NodeServer = &NodeServer{}
const protocolFileName string = `protocolConn.json`

func NewNodeServer(nodeId, driverName string, c cs.ObjectstorageV1alpha1Client) csi.NodeServer {
	return &NodeServer{
		name:       driverName,
		nodeID:     nodeId,
		cosiClient: c,
		ctx:        context.Background(),
	}
}


// logErr should be called at the interface method scope, prior to returning errors to the gRPC client.
func logErr(e error) error {
	klog.Error(e)
	return e
}

// NodeServer implements the NodePublishVolume and NodeUnpublishVolume methods
// of the csi.NodeServer interface and GetPluginCapabilities, GetPluginInfo, and
// Probe of the IdentityServer interface.
type NodeServer struct {
	name       string
	nodeID     string
	cosiClient cs.ObjectstorageV1alpha1Client
	ctx        context.Context
}

func (n NodeServer) NodeStageVolume(ctx context.Context, request *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	panic("implement me")
}

func (n NodeServer) NodeUnstageVolume(ctx context.Context, request *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	panic("implement me")
}


const (
	barName      = "bucketAccessRequestName"
	barNamespace = "bucketAccessRequestNamespace"
)

func bucketAccessRequestNameNamespace(volCtx map[string]string) (name, ns string, err error) {
	klog.Info("parsing bucketAccessRequest namespace/name from volume context")
	e := func(m string) error { return fmt.Errorf("required volume context key unset: %v", m) }

	var ok bool
	name, ok = volCtx[barName]
	if ! ok {
		return "", "", e(barName)
	}
	klog.Infof("got name: %v", name)
	ns, ok = volCtx[barNamespace]
	if ! ok {
		return "", "", e(barNamespace)
	}
	klog.Infof("got namespace: %v", ns)
	return
}

func (n NodeServer) NodePublishVolume(ctx context.Context, request *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.Infof("NodePublishVolume: volId: %v, targetPath: %v\n", request.GetVolumeId(), request.GetTargetPath())

	name, ns, err := bucketAccessRequestNameNamespace(request.VolumeContext)
	if err != nil {
		return nil, err
	}

	getError := func(t, n string, e error) error { return fmt.Errorf("failed to get <%s>%s: %v", t, n, e) }

	klog.Infof("getting bucketAccessRequest %q", fmt.Sprintf("%s/%s", ns, name))
	bar, err := n.cosiClient.BucketAccessRequests(ns).Get(n.ctx, name, v1.GetOptions{})
	if err != nil || bar == nil {
		return nil, logErr(getError("bucketAccessRequest", fmt.Sprintf("%s/%s", ns, name), err))
	}
	if len(bar.Spec.BucketRequestName) == 0 {
		return nil, logErr(fmt.Errorf("bucketAccessRequest.Spec.BucketRequestName unset"))
	}

	klog.Infof("getting bucketRequest %q", bar.Spec.BucketRequestName)
	br, err := n.cosiClient.BucketRequests(bar.Namespace).Get(n.ctx, bar.Spec.BucketRequestName, v1.GetOptions{})
	if err != nil || br == nil {
		return nil, logErr(getError("bucketRequest", fmt.Sprintf("%s/%s", bar.Namespace, bar.Spec.BucketRequestName), err))
	}

	klog.Infof("getting bucket %q", br.Spec.BucketInstanceName)
	// is BucketInstanceName the correct field, or should it be BucketClass
	bkt, err := n.cosiClient.Buckets().Get(n.ctx, br.Spec.BucketInstanceName, v1.GetOptions{})
	if err != nil || bkt == nil {
		return nil, logErr(getError("bucket", br.Spec.BucketInstanceName, err))
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
		err = fmt.Errorf("bucket %q protocol not signature")
	default:
		err = fmt.Errorf("unrecognized protocol %q, unable to extract connection data", bkt.Spec.Protocol)
	}
	if err != nil {
		return nil, logErr(err)
	}
	klog.Infof("bucket %q has protocol %q", bkt.Name, bkt.Spec.Protocol)

	protoData, err := json.Marshal(protocolConnection)
	if err != nil {
		return nil, logErr(fmt.Errorf("error marshalling protocol: %v", err))
	}

	target := filepath.Join(request.TargetPath, protocolFileName)
	klog.Infof("creating conn file: %s", target)
	f, err := os.Open(target)
	if err != nil {
		return nil, logErr(fmt.Errorf("error creating file: %s: %v", target, err))
	}
	defer f.Close()
	_, err = f.Write(protoData)
	if err != nil {
		return nil, logErr(fmt.Errorf("unable to write to file: %v", err))
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

func (n NodeServer) NodeUnpublishVolume(ctx context.Context, request *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.Infof("NodeUnpublishVolume: volId: %v, targetPath: %v\n", request.GetVolumeId(), request.GetTargetPath())
	target := filepath.Join(request.TargetPath, protocolFileName)
	err := os.RemoveAll(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, logErr(fmt.Errorf("unable to remove file %s: %v", target, err))
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (n NodeServer) NodeGetVolumeStats(ctx context.Context, request *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	panic("implement me")
}

func (n NodeServer) NodeExpandVolume(ctx context.Context, request *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	panic("implement me")
}

func (n NodeServer) NodeGetCapabilities(ctx context.Context, request *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	panic("implement me")
}

func (n NodeServer) NodeGetInfo(ctx context.Context, request *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.Infof("NodeGetInfo()")
	resp := &csi.NodeGetInfoResponse{
		NodeId: n.nodeID,
	}
	return resp, nil
}

