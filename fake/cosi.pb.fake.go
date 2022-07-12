package fake

import (
	"context"
	grpc "google.golang.org/grpc"
	containerobjectstorageinterfacespec "sigs.k8s.io/container-object-storage-interface-spec"
)

type FakeIdentityClient struct {
	FakeDriverGetInfo func(ctx context.Context, in *containerobjectstorageinterfacespec.DriverGetInfoRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverGetInfoResponse, error)
}

func (f *FakeIdentityClient) DriverGetInfo(ctx context.Context, in *containerobjectstorageinterfacespec.DriverGetInfoRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverGetInfoResponse, error) {
	return f.FakeDriverGetInfo(ctx, in, opts...)
}

type FakeProvisionerClient struct {
	FakeDriverCreateBucket       func(ctx context.Context, in *containerobjectstorageinterfacespec.DriverCreateBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverCreateBucketResponse, error)
	FakeDriverDeleteBucket       func(ctx context.Context, in *containerobjectstorageinterfacespec.DriverDeleteBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverDeleteBucketResponse, error)
	FakeDriverGrantBucketAccess  func(ctx context.Context, in *containerobjectstorageinterfacespec.DriverGrantBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverGrantBucketAccessResponse, error)
	FakeDriverRevokeBucketAccess func(ctx context.Context, in *containerobjectstorageinterfacespec.DriverRevokeBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverRevokeBucketAccessResponse, error)
}

func (f *FakeProvisionerClient) DriverCreateBucket(ctx context.Context, in *containerobjectstorageinterfacespec.DriverCreateBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverCreateBucketResponse, error) {
	return f.FakeDriverCreateBucket(ctx, in, opts...)
}
func (f *FakeProvisionerClient) DriverDeleteBucket(ctx context.Context, in *containerobjectstorageinterfacespec.DriverDeleteBucketRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverDeleteBucketResponse, error) {
	return f.FakeDriverDeleteBucket(ctx, in, opts...)
}
func (f *FakeProvisionerClient) DriverGrantBucketAccess(ctx context.Context, in *containerobjectstorageinterfacespec.DriverGrantBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverGrantBucketAccessResponse, error) {
	return f.FakeDriverGrantBucketAccess(ctx, in, opts...)
}
func (f *FakeProvisionerClient) DriverRevokeBucketAccess(ctx context.Context, in *containerobjectstorageinterfacespec.DriverRevokeBucketAccessRequest, opts ...grpc.CallOption) (*containerobjectstorageinterfacespec.DriverRevokeBucketAccessResponse, error) {
	return f.FakeDriverRevokeBucketAccess(ctx, in, opts...)
}
