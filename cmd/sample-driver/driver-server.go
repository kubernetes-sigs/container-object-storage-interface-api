/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/minio/minio-go"
	"github.com/minio/minio/pkg/auth"
	"github.com/minio/minio/pkg/bucket/policy"
	"github.com/minio/minio/pkg/bucket/policy/condition"
	iampolicy "github.com/minio/minio/pkg/iam/policy"
	"github.com/minio/minio/pkg/madmin"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"

	cosi "github.com/kubernetes-sigs/container-object-storage-interface-spec"
)

var (
	PROVISIONER_NAME = "sample-provisioner.objectstorage.k8s.io"
	VERSION          = "dev"
)

type DriverServer struct {
	Name, Version string
	S3Client      *minio.Client
	S3AdminClient *madmin.AdminClient
}

func (ds *DriverServer) ProvisionerGetInfo(context.Context, *cosi.ProvisionerGetInfoRequest) (*cosi.ProvisionerGetInfoResponse, error) {
	rsp := &cosi.ProvisionerGetInfoResponse{}
	rsp.Name = fmt.Sprintf("%s-%s", ds.Name, ds.Version)
	return rsp, nil
}

func (ds DriverServer) ProvisionerCreateBucket(ctx context.Context, req *cosi.ProvisionerCreateBucketRequest) (*cosi.ProvisionerCreateBucketResponse, error) {
	klog.Infof("Using minio to create Backend Bucket")

	if ds.Name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if ds.Version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	err := ds.S3Client.MakeBucket(req.BucketName, "")
	if err != nil {
		// Check to see if the bucket already exists
		exists, errBucketExists := ds.S3Client.BucketExists(req.BucketName)
		if errBucketExists == nil && exists {
			klog.Info("Backend Bucket already exists", req.BucketName)
			return &cosi.ProvisionerCreateBucketResponse{}, nil
		} else {
			klog.Error(err)
			return &cosi.ProvisionerCreateBucketResponse{}, err
		}
	}
	klog.Info("Successfully created Backend Bucket", req.BucketName)

	return &cosi.ProvisionerCreateBucketResponse{}, nil
}

func (ds *DriverServer) ProvisionerDeleteBucket(ctx context.Context, req *cosi.ProvisionerDeleteBucketRequest) (*cosi.ProvisionerDeleteBucketResponse, error) {

	if err := ds.S3Client.RemoveBucket(req.BucketName); err != nil {
		klog.Info("failed to delete bucket", req.BucketName)
		return nil, err

	}
	return &cosi.ProvisionerDeleteBucketResponse{}, nil
}

func (ds *DriverServer) ProvisionerGrantBucketAccess(ctx context.Context, req *cosi.ProvisionerGrantBucketAccessRequest) (*cosi.ProvisionerGrantBucketAccessResponse, error) {

	creds, err := auth.GetNewCredentials()
	if err != nil {
		klog.Error("failed to generate new credentails")
		return nil, err
	}

	if err := ds.S3AdminClient.AddUser(context.Background(), creds.AccessKey, creds.SecretKey); err != nil {
		klog.Error("failed to create user", err)
		return nil, err
	}

	// Create policy
	p := iampolicy.Policy{
		Version: iampolicy.DefaultVersion,
		Statements: []iampolicy.Statement{
			iampolicy.NewStatement(
				policy.Allow,
				iampolicy.NewActionSet("s3:*"),
				iampolicy.NewResourceSet(iampolicy.NewResource(req.GetBucketName()+"/*", "")),
				condition.NewFunctions(),
			)},
	}

	if err := ds.S3AdminClient.AddCannedPolicy(context.Background(), "s3:*", &p); err != nil {
		klog.Error("failed to add canned policy", err)
		return nil, err
	}

	if err := ds.S3AdminClient.SetPolicy(context.Background(), "s3:*", creds.AccessKey, false); err != nil {
		klog.Error("failed to set policy", err)
		return nil, err
	}

	return &cosi.ProvisionerGrantBucketAccessResponse{
		Principal:               req.Principal,
		CredentialsFileContents: fmt.Sprintf("[default]\naws_access_key %s\naws_secret_key %s", creds.AccessKey, creds.SecretKey),
		CredentialsFilePath:     ".aws/credentials",
	}, nil
}

func (ds *DriverServer) ProvisionerRevokeBucketAccess(ctx context.Context, req *cosi.ProvisionerRevokeBucketAccessRequest) (*cosi.ProvisionerRevokeBucketAccessResponse, error) {

	// revokes user access to bucket
	if err := ds.S3AdminClient.RemoveUser(ctx, req.GetPrincipal()); err != nil {
		klog.Error("falied to Revoke Bucket Access")
		return nil, err
	}
	return &cosi.ProvisionerRevokeBucketAccessResponse{}, nil
}
