module github.com/container-object-storage-interface/ephemeral-csi-driver

go 1.15

require (
	github.com/container-object-storage-interface/api v0.0.0-20200919075357-9eea1b2d66da
	github.com/container-object-storage-interface/spec v0.0.0-20200908192509-18912d8bf2a5
	github.com/container-storage-interface/spec v1.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	google.golang.org/grpc v1.30.0
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v0.18.4
	k8s.io/klog v1.0.0
)
