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

package internal

import (
	"context"

	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/cmd/minio-cosi-driver/internal/minio"
)

func NewDriver(ctx context.Context, provisioner, minioHost, accessKey, secretKey string) (*IdentityServer, *ProvisionerServer, error) {
	mc, err := minio.NewClient(ctx, minioHost, accessKey, secretKey)
	if err != nil {
		return nil, nil, err
	}

	return &IdentityServer{
			provisioner: provisioner,
		}, &ProvisionerServer{
			provisioner: provisioner,
			mc:          mc,
		}, nil
}
