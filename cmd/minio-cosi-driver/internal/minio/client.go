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
	"net/url"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	min "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"k8s.io/klog/v2"
)

type C struct {
	accessKey string
	secretKey string
	host      *url.URL

	client *min.Client
}

func NewClient(ctx context.Context, minioHost, accessKey, secretKey string) (*C, error) {
	if minioHost == "" {
		return nil, errors.New("minio host cannot be empty")
	}
	host, err := url.Parse(minioHost)
	if err != nil {
		return nil, err
	}

	secure := false
	switch host.Scheme {
	case "http":
	case "https":
		secure = true
	default:
		return nil, errors.New("invalid url scheme for minio endpoint")
	}

	clChan := make(chan *min.Client)
	errChan := make(chan error)
	go func() {
		klog.V(3).InfoS("Connecting to MinIO", "endpoint", host.Host)

		cl, err := min.New(host.Host, &min.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: secure,
		})
		if err != nil {
			errChan <- err
		}
		_, err = cl.BucketExists(ctx, uuid.New().String())
		if err != nil {
			if errResp, ok := err.(min.ErrorResponse); ok {
				if errResp.Code == "NoSuchBucket" {
					clChan <- cl
					return
				}
				if errResp.StatusCode == 403 {
					errChan <- errors.Wrap(errors.New("Access Denied"), "Connection to MinIO Failed")
					return
				}
			}
			errChan <- errors.Wrap(err, "Connection to MinIO Failed")
			return
		}

		clChan <- cl
		klog.InfoS("Successfully connected to MinIO")
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case cl := <-clChan:
		return &C{
			accessKey: accessKey,
			secretKey: secretKey,
			host:      host,

			client: cl,
		}, nil
	case err := <-errChan:
		return nil, err
	}
}
