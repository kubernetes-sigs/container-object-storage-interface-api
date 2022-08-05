module sigs.k8s.io/container-object-storage-interface-provisioner-sidecar

go 1.15

require (
	github.com/google/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.12.0
	google.golang.org/grpc v1.46.2
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
	k8s.io/client-go v0.24.2
	k8s.io/klog/v2 v2.70.1
	sigs.k8s.io/container-object-storage-interface-api v0.0.0-20220727205553-02ff3dd25b5e
	sigs.k8s.io/container-object-storage-interface-spec v0.0.0-20220804173401-3154aa8927e3
	sigs.k8s.io/controller-runtime v0.12.3
)
