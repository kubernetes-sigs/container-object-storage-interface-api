package app

import (
	"context"
	"os"
	"time"

	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/controller/bucket"
	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/controller/bucketaccess"
	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/grpcclient"

	osspec "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/spf13/cobra"

	"google.golang.org/grpc"

	"k8s.io/klog/v2"
)

const (
	// Interval of logging connection errors
	connectionLoggingInterval = 10 * time.Second
	defaultDriverAddress      = "tcp://0.0.0.0:9000"
)

// SidecarOptions defines the options for running the sidecar
type SidecarOptions struct {
	driverAddress string
}

// NewSidecarOptions returns an initialized SidecarOptions instance
func NewSidecarOptions() *SidecarOptions {
	return &SidecarOptions{driverAddress: defaultDriverAddress}
}

// Run starts the sidecar with the configured options
func (so *SidecarOptions) Run() {
	klog.Infof("attempting to open a gRPC connection with: %q", so.driverAddress)
	grpcClient, err := grpcclient.NewGRPCClient(so.driverAddress, []grpc.DialOption{}, nil)
	if err != nil {
		klog.Errorf("error creating GRPC Client: %v", err)
		os.Exit(1)
	}

	grpcConn, err := grpcClient.ConnectWithLogging(connectionLoggingInterval)
	if err != nil {
		klog.Errorf("error connecting to COSI driver: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	klog.Infof("creating provisioner client")
	provisionerClient := osspec.NewProvisionerClient(grpcConn)

	klog.Infof("discovering driver name")
	req := osspec.ProvisionerGetInfoRequest{}
	rsp, err := provisionerClient.ProvisionerGetInfo(ctx, &req)
	if err != nil {
		klog.Errorf("error calling ProvisionerGetInfo: %v", err)
		os.Exit(1)
	}

	provisionerName := rsp.Name
	// TODO: Register provisioner using internal type
	klog.Info("This sidecar is working with the driver identified as: ", provisionerName)

	so.startControllers(ctx, provisionerName, provisionerClient)
	<-ctx.Done()
}

func (so *SidecarOptions) startControllers(ctx context.Context, name string, client osspec.ProvisionerClient) {
	bucketController, err := bucket.NewBucketController(name, client)
	if err != nil {
		klog.Fatalf("Error creating bucket controller: %v", err)
	}

	bucketAccessController, err := bucketaccess.NewBucketAccessController(name, client)
	if err != nil {
		klog.Fatalf("Error creating bucket access controller: %v", err)
	}

	go bucketController.Run(ctx)
	go bucketAccessController.Run(ctx)
}

// NewControllerManagerCommand creates a *cobra.Command object with default parameters
func NewControllerManagerCommand() *cobra.Command {
	opts := NewSidecarOptions()

	cmd := &cobra.Command{
		Use:                   "objectstorage-sidecar",
		DisableFlagsInUseLine: true,
		Short:                 "",
		Long:                  ``,
		Run: func(cmd *cobra.Command, args []string) {
			opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.driverAddress, "connect-address", "c", opts.driverAddress, "The address that the sidecar should connect to")

	return cmd
}
