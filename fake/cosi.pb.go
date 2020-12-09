package fake

import (
	"context"

	cosi "github.com/kubernetes-sigs/container-object-storage-interface-spec"

	"google.golang.org/grpc"
)

// this ensures that the mock implements the client interface
var _ cosi.ProvisionerClient = (*MockProvisionerClient)(nil)

// MockProvisionerClient is a type that implements all the methods for RolePolicyAttachmentClient interface
type MockProvisionerClient struct {
	GetInfo            func(ctx context.Context, in *cosi.ProvisionerGetInfoRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGetInfoResponse, error)
	CreateBucket       func(ctx context.Context, in *cosi.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerCreateBucketResponse, error)
	DeleteBucket       func(ctx context.Context, in *cosi.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error)
	GrantBucketAccess  func(ctx context.Context, in *cosi.ProvisionerGrantBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error)
	RevokeBucketAccess func(ctx context.Context, in *cosi.ProvisionerRevokeBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerRevokeBucketAccessResponse, error)
}

// ProvisionerCreateBucket mocks GetBucketPolicyRequest method
func (m *MockProvisionerClient) ProvisionerCreateBucket(ctx context.Context, in *cosi.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerCreateBucketResponse, error) {
	return m.CreateBucket(ctx, in, opts...)
}

// ProvisionerDeleteBucket mocks PutBucketPolicyRequest method
func (m *MockProvisionerClient) ProvisionerDeleteBucket(ctx context.Context, in *cosi.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error) {
	return m.DeleteBucket(ctx, in, opts...)
}

// ProvisionerGrantBucketAccess mocks DeleteBucketPolicyRequest method
func (m *MockProvisionerClient) ProvisionerGrantBucketAccess(ctx context.Context, in *cosi.ProvisionerGrantBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error) {
	return m.GrantBucketAccess(ctx, in, opts...)
}

// ProvisionerRevokeBucketAccess mocks DeleteBucketPolicyRequest method
func (m *MockProvisionerClient) ProvisionerRevokeBucketAccess(ctx context.Context, in *cosi.ProvisionerRevokeBucketAccessRequest, opts ...grpc.CallOption) (*cosi.ProvisionerRevokeBucketAccessResponse, error) {
	return m.RevokeBucketAccess(ctx, in, opts...)
}

// ProvisionerRevokeBucketAccess mocks DeleteBucketPolicyRequest method
func (m *MockProvisionerClient) ProvisionerGetInfo(ctx context.Context, in *cosi.ProvisionerGetInfoRequest, opts ...grpc.CallOption) (*cosi.ProvisionerGetInfoResponse, error) {
	return m.GetInfo(ctx, in, opts...)
}
