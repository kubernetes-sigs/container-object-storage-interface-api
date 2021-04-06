/* Copyright 2021 The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package provisioner

import (
	"context"

	"google.golang.org/grpc"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

var (
	_ cosi.IdentityClient    = &COSIProvisionerClient{}
	_ cosi.ProvisionerClient = &COSIProvisionerClient{}
)

type COSIProvisionerClient struct {
	address           string
	conn              *grpc.ClientConn
	identityClient    cosi.IdentityClient
	provisionerClient cosi.ProvisionerClient
}

func (c *COSIProvisionerClient) ProvisionerGetInfo(ctx context.Context,
	in *cosi.ProvisionerGetInfoRequest,
	opts ...grpc.CallOption) (*cosi.ProvisionerGetInfoResponse, error) {

	return c.identityClient.ProvisionerGetInfo(ctx, in, opts...)
}

func (c *COSIProvisionerClient) ProvisionerCreateBucket(ctx context.Context,
	in *cosi.ProvisionerCreateBucketRequest,
	opts ...grpc.CallOption) (*cosi.ProvisionerCreateBucketResponse, error) {

	return c.provisionerClient.ProvisionerCreateBucket(ctx, in, opts...)
}

func (c *COSIProvisionerClient) ProvisionerDeleteBucket(ctx context.Context,
	in *cosi.ProvisionerDeleteBucketRequest,
	opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error) {

	return c.provisionerClient.ProvisionerDeleteBucket(ctx, in, opts...)
}

func (c *COSIProvisionerClient) ProvisionerGrantBucketAccess(ctx context.Context,
	in *cosi.ProvisionerGrantBucketAccessRequest,
	opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error) {

	return c.provisionerClient.ProvisionerGrantBucketAccess(ctx, in, opts...)
}

func (c *COSIProvisionerClient) ProvisionerRevokeBucketAccess(ctx context.Context,
	in *cosi.ProvisionerRevokeBucketAccessRequest,
	opts ...grpc.CallOption) (*cosi.ProvisionerRevokeBucketAccessResponse, error) {

	return c.provisionerClient.ProvisionerRevokeBucketAccess(ctx, in, opts...)
}
