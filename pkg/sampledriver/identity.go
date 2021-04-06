// Copyright 2021 The Kubernetes Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sampledriver

import (
	"context"

	"k8s.io/klog/v2"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IdentityServer struct {
	provisioner string
}

func (id *IdentityServer) ProvisionerGetInfo(ctx context.Context,
	req *cosi.ProvisionerGetInfoRequest) (*cosi.ProvisionerGetInfoResponse, error) {

	if id.provisioner == "" {
		klog.ErrorS(errors.New("provisioner name cannot be empty"), "Invalid argument")
		return nil, status.Error(codes.Unavailable, "Provisioner name not configured")
	}

	return &cosi.ProvisionerGetInfoResponse{
		Name: id.provisioner,
	}, nil
}
