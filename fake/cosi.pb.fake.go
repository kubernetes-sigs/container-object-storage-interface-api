package fake

import (
	"context"
	"google.golang.org/grpc"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

type FakeIdentityClient struct {
	FakeProvisionerGetInfo func(ctx context.Context, in *cosi.ProvisionerGetInfoRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGetInfoResponse, error)
}

func (f *FakeIdentityClient) ProvisionerGetInfo(ctx context.Context, in *cosi.ProvisionerGetInfoRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGetInfoResponse, error) {
	return f.FakeProvisionerGetInfo(ctx, in, opts...)
}

type FakeProvisionerClient struct {
	FakeProvisionerCreateBucket func(ctx context.Context, in *cosi.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerCreateBucketResponse, error)
	FakeProvisionerDeleteBucket func(ctx context.Context, in *cosi.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error)
	FakeProvisionerGrantBucketAccess func(ctx context.Context, in *cosi.ProvisionerGrantBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error)
	FakeProvisionerRevokeBucketAccess func(ctx context.Context, in *cosi.ProvisionerRevokeBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerRevokeBucketAccessResponse, error)
}

func (f *FakeProvisionerClient) ProvisionerCreateBucket(ctx context.Context, in *cosi.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerCreateBucketResponse, error) {
	return f.FakeProvisionerCreateBucket(ctx, in, opts...)
}

func (f *FakeProvisionerClient) ProvisionerDeleteBucket(ctx context.Context, in *cosi.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error) {
	return f.FakeProvisionerDeleteBucket(ctx, in, opts...)
}

func (f *FakeProvisionerClient) ProvisionerGrantBucketAccess(ctx context.Context, in *cosi.ProvisionerGrantBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error) {
	return f.FakeProvisionerGrantBucketAccess(ctx, in, opts...)
}

func (f *FakeProvisionerClient) ProvisionerRevokeBucketAccess(ctx context.Context, in *cosi.ProvisionerRevokeBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerRevokeBucketAccessResponse, error) {
	return f.FakeProvisionerRevokeBucketAccess(ctx, in, opts...)
}

