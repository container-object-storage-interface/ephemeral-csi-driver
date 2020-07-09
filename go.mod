module cosi-csi-driver

go 1.13

require (
	github.com/container-object-storage-interface/api v0.0.0-20200627005153-101660685c39
	github.com/container-storage-interface/spec v1.3.0
	github.com/golang/protobuf v1.4.2 // indirect
	google.golang.org/grpc v1.29.1
	k8s.io/apimachinery v0.18.4
	k8s.io/klog v1.0.0
)

replace github.com/container-object-storage-interface/api => /Users/jcope/Workspace/go/src/github.com/container-object-storage-interface/api
