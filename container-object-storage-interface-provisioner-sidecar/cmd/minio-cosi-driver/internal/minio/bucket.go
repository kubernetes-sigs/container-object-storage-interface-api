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

package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

var ErrBucketAlreadyExists = errors.New("Bucket Already Exists")

type MakeBucketOptions minio.MakeBucketOptions

func (x *C) CreateBucket(ctx context.Context, bucketName string, options MakeBucketOptions) (string, error) {
	if err := x.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions(options)); err != nil {
		errCode := minio.ToErrorResponse(err).Code
		if errCode == "BucketAlreadyExists" || errCode == "BucketAlreadyOwnedByYou" {
			return bucketName, ErrBucketAlreadyExists
		}
		return "", err
	}
	return bucketName, nil
}
