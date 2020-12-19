module github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar

go 1.15

require (
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/kubernetes-sigs/container-object-storage-interface-api v0.0.0-20201217233824-6b4158ff7e28
	github.com/kubernetes-sigs/container-object-storage-interface-spec v0.0.0-20201217184109-8cbf84dde8d3
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324
	google.golang.org/grpc v1.34.0
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/klog v1.0.0
)
