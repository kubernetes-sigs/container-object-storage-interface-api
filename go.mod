module github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar

go 1.14

require (
	github.com/kubernetes-sigs/container-object-storage-interface-api v0.0.0-20201204201926-43539346a903
	github.com/kubernetes-sigs/container-object-storage-interface-spec v0.0.0-20201208142312-59e00cb00687
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324
	google.golang.org/grpc v1.34.0
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/klog v1.0.0
)
