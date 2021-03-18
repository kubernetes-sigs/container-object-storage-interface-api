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

	"k8s.io/klog/v2"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

type DriverServer struct {
	S3Client      *minio.Client
	S3AdminClient *madmin.AdminClient
}

func (ds DriverServer) ProvisionerCreateBucket(ctx context.Context, req *cosi.ProvisionerCreateBucketRequest) (*cosi.ProvisionerCreateBucketResponse, error) {
	klog.Infof("Using minio to create Backend Bucket")

	s3 := req.Protocol.GetS3()
	if s3 == nil {
		return nil, status.Error(codes.Unavailable, "Driver is missing protocol")
	}

	err := ds.S3Client.MakeBucket(s3.BucketName, "")
	if err != nil {
		// Check to see if the bucket already exists
		exists, errBucketExists := ds.S3Client.BucketExists(s3.BucketName)
		if errBucketExists == nil && exists {
			klog.Info("Backend Bucket already exists", s3.BucketName)
			return &cosi.ProvisionerCreateBucketResponse{}, nil
		} else {
			klog.Error(err)
			return &cosi.ProvisionerCreateBucketResponse{}, err
		}
	}
	klog.Info("Successfully created Backend Bucket", s3.BucketName)

	return &cosi.ProvisionerCreateBucketResponse{}, nil
}

func (ds *DriverServer) ProvisionerDeleteBucket(ctx context.Context, req *cosi.ProvisionerDeleteBucketRequest) (*cosi.ProvisionerDeleteBucketResponse, error) {
	s3 := req.Protocol.GetS3()
	if s3 == nil {
		return nil, status.Error(codes.Unavailable, "Driver is missing protocol")
	}

	if err := ds.S3Client.RemoveBucket(s3.BucketName); err != nil {
		klog.Info("failed to delete bucket", s3.BucketName)
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

	s3 := req.Protocol.GetS3()
	if s3 == nil {
		return nil, status.Error(codes.Unavailable, "Driver is missing protocol")
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
				iampolicy.NewResourceSet(iampolicy.NewResource(s3.BucketName+"/*", "")),
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
