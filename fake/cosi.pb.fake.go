package fake

import (
	"context"
	grpc "google.golang.org/grpc"
	containerobjectstorageinterfacespec "sigs.k8s.io/container-object-storage-interface-spec"
)

type FakeIdentityClient struct {
	FakeProvisionerGetInfo func(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerGetInfoRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerGetInfoResponse, error)
}

func (f *FakeIdentityClient) ProvisionerGetInfo(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerGetInfoRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerGetInfoResponse, error) {
	return f.FakeProvisionerGetInfo(ctx, in, opts...)
}

type FakeProvisionerClient struct {
	FakeProvisionerCreateBucket       func(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerCreateBucketResponse, error)
	FakeProvisionerDeleteBucket       func(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerDeleteBucketResponse, error)
	FakeProvisionerGrantBucketAccess  func(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerGrantBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerGrantBucketAccessResponse, error)
	FakeProvisionerRevokeBucketAccess func(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerRevokeBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerRevokeBucketAccessResponse, error)
}

func (f *FakeProvisionerClient) ProvisionerCreateBucket(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerCreateBucketResponse, error) {
	return f.FakeProvisionerCreateBucket(ctx, in, opts...)
}
func (f *FakeProvisionerClient) ProvisionerDeleteBucket(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerDeleteBucketResponse, error) {
	return f.FakeProvisionerDeleteBucket(ctx, in, opts...)
}
func (f *FakeProvisionerClient) ProvisionerGrantBucketAccess(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerGrantBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerGrantBucketAccessResponse, error) {
	return f.FakeProvisionerGrantBucketAccess(ctx, in, opts...)
}
func (f *FakeProvisionerClient) ProvisionerRevokeBucketAccess(ctx context.Context, in *containerobjectstorageinterfacespec.ProvisionerRevokeBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.ProvisionerRevokeBucketAccessResponse, error) {
	return f.FakeProvisionerRevokeBucketAccess(ctx, in, opts...)
}
