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
	"github.com/minio/minio/pkg/madmin"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

var (
	PROVISIONER_NAME = "sample-provisioner.objectstorage.k8s.io"
	VERSION          = "dev"
)

type IdentityServer struct {
	Name, Version string
	S3Client      *minio.Client
	S3AdminClient *madmin.AdminClient
}

func (id *IdentityServer) ProvisionerGetInfo(context.Context, *cosi.ProvisionerGetInfoRequest) (*cosi.ProvisionerGetInfoResponse, error) {
	if id.Name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if id.Version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}
	rsp := &cosi.ProvisionerGetInfoResponse{}
	rsp.Name = fmt.Sprintf("%s-%s", id.Name, id.Version)
	return rsp, nil
}
