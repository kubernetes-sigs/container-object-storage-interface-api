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

package bucket

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1"
	buckets "sigs.k8s.io/container-object-storage-interface-api/clientset"
	bucketapi "sigs.k8s.io/container-object-storage-interface-api/clientset/typed/objectstorage.k8s.io/v1alpha1"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BucketListener manages Bucket objects
type BucketListener struct {
	provisionerClient cosi.ProvisionerClient
	provisionerName   string

	kubeClient   kube.Interface
	bucketClient buckets.Interface
	kubeVersion  *utilversion.Version
}

// NewBucketListener returns a resource handler for Bucket objects
func NewBucketListener(provisionerName string, client cosi.ProvisionerClient) *BucketListener {
	bl := &BucketListener{
		provisionerName:   provisionerName,
		provisionerClient: client,
	}

	return bl
}

// Add attempts to create a bucket for a given bucket. This function must be idempotent
// Return values
//    nil - Bucket successfully provisioned
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (b *BucketListener) Add(ctx context.Context, inputBucket *v1alpha1.Bucket) error {
	bucket := inputBucket.DeepCopy()

	klog.V(3).InfoS("Add Bucket",
		"name", bucket.Name,
		"bucketclass", bucket.Spec.BucketClassName,
	)

	if !strings.EqualFold(bucket.Spec.Provisioner, b.provisionerName) {
		klog.V(5).InfoS("Skipping bucket for provisiner",
			"bucket", bucket.Name,
			"provisioner", bucket.Spec.Provisioner,
		)
		return nil
	}

	if bucket.Status.BucketAvailable {
		klog.V(5).InfoS("BucketExists",
			"bucket", bucket.Name,
			"provisioner", bucket.Spec.Provisioner,
		)
		return nil
	}

	proto, err := bucket.Spec.Protocol.ConvertToExternal()
	if err != nil {
		klog.ErrorS(err, "Invalid protocol",
			"bucket", bucket.Name)

		return errors.Wrap(err, "Failed to parse protocol for API")
	}

	req := &cosi.ProvisionerCreateBucketRequest{
		Parameters: bucket.Spec.Parameters,
		Protocol:   proto,
		Name:       bucket.Name,
	}

	rsp, err := b.provisionerClient.ProvisionerCreateBucket(ctx, req)
	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			klog.ErrorS(err, "Failed to create bucket",
				"bucket", bucket.Name)
			return errors.Wrap(err, "Failed to create bucket")
		}
	}
	if rsp == nil {
		err := errors.New("ProvisionerCreateBucket returned a nil response")
		klog.ErrorS(err, "Internal Error")
		return err
	}

	if rsp.BucketId != "" {
		bucket.Spec.BucketID = rsp.BucketId
	}
	bucket.Status.Message = "Bucket Provisioned"
	bucket.Status.BucketAvailable = true

	// if this step fails, then controller will retry with backoff
	if _, err := b.Buckets().Update(ctx, bucket, metav1.UpdateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update bucket",
			"bucket", bucket.Name)
		return errors.Wrap(err, "Failed to update bucket")
	}

	return nil
}

// Update attempts to reconcile changes to a given bucket. This function must be idempotent
// Return values
//    nil - Bucket successfully reconciled
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (b *BucketListener) Update(ctx context.Context, old, new *v1alpha1.Bucket) error {
	klog.V(3).InfoS("Update Bucket",
		"name", old.Name)

	return nil
}

// Delete attemps to delete a bucket. This function must be idempotent
// Return values
//    nil - Bucket successfully deleted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (b *BucketListener) Delete(ctx context.Context, inputBucket *v1alpha1.Bucket) error {
	bucket := inputBucket.DeepCopy()

	klog.V(3).InfoS("Delete Bucket",
		"name", bucket.Name,
		"bucketclass", bucket.Spec.BucketClassName,
	)

	if !strings.EqualFold(bucket.Spec.Provisioner, b.provisionerName) {
		klog.V(5).InfoS("Skipping bucket for provisiner",
			"bucket", bucket.Name,
			"provisioner", bucket.Spec.Provisioner,
		)
		return nil
	}

	req := &cosi.ProvisionerDeleteBucketRequest{
		BucketId: bucket.Spec.BucketID,
	}

	if _, err := b.provisionerClient.ProvisionerDeleteBucket(ctx, req); err != nil {
		if status.Code(err) != codes.NotFound {
			klog.ErrorS(err, "Failed to delete bucket",
				"bucket", bucket.Name,
			)
			return err
		}
	}

	bucket.Status.BucketAvailable = false

	// if this step fails, then controller will retry with backoff
	if _, err := b.Buckets().Update(ctx, bucket, metav1.UpdateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update bucket",
			"bucket", bucket.Name)
		return errors.Wrap(err, "Failed to update bucket")
	}

	return nil
}

func (b *BucketListener) Buckets() bucketapi.BucketInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().Buckets()
	}
	panic("uninitialized listener")
}

// InitializeKubeClient initializes the kubernetes client
func (b *BucketListener) InitializeKubeClient(k kube.Interface) {
	b.kubeClient = k

	serverVersion, err := k.Discovery().ServerVersion()
	if err != nil {
		klog.ErrorS(err, "Cannot determine server version")
	} else {
		b.kubeVersion = utilversion.MustParseSemantic(serverVersion.GitVersion)
	}
}

// InitializeBucketClient initializes the object storage bucket client
func (b *BucketListener) InitializeBucketClient(bc buckets.Interface) {
	b.bucketClient = bc
}
