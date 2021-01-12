module sigs.k8s.io/container-object-storage-interface-provisioner-sidecar

go 1.15

require (
	github.com/go-ini/ini v1.62.0 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/kubernetes-sigs/container-object-storage-interface-api v0.0.0-20210112203836-a74c106d3ab3
	github.com/kubernetes-sigs/container-object-storage-interface-spec v0.0.0-20210107181053-ef523592ca74
	github.com/minio/minio v0.0.0-20210112204746-e09196d62633
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	golang.org/x/net v0.0.0-20201216054612-986b41b23924
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324
	google.golang.org/grpc v1.34.0
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/klog v1.0.0
)
