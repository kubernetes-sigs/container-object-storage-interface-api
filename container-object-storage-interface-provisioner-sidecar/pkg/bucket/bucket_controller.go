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

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	buckets "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned"
	bucketapi "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned/typed/objectstorage/v1alpha1"
	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/consts"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BucketListener manages Bucket objects
type BucketListener struct {
	provisionerClient cosi.ProvisionerClient
	driverName        string

	kubeClient   kube.Interface
	bucketClient buckets.Interface
	kubeVersion  *utilversion.Version
}

// NewBucketListener returns a resource handler for Bucket objects
func NewBucketListener(driverName string, client cosi.ProvisionerClient) *BucketListener {
	bl := &BucketListener{
		driverName:        driverName,
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

	var err error

	klog.V(3).InfoS("Add Bucket",
		"name", bucket.ObjectMeta.Name,
		"bucketclass", bucket.Spec.BucketClassName,
	)

	if !strings.EqualFold(bucket.Spec.DriverName, b.driverName) {
		klog.V(5).InfoS("Skipping bucket for driver",
			"bucket", bucket.ObjectMeta.Name,
			"driver", bucket.Spec.DriverName,
		)
		return nil
	}

	if bucket.Status.BucketReady {
		klog.V(5).InfoS("BucketExists",
			"bucket", bucket.ObjectMeta.Name,
			"driver", bucket.Spec.DriverName,
		)

		return nil
	}

	if bucket.Spec.ExistingBucketID != "" {
		bucket.Status.BucketReady = true
		bucket.Status.BucketID = bucket.Spec.ExistingBucketID

	} else {
		req := &cosi.DriverCreateBucketRequest{
			Parameters: bucket.Spec.Parameters,
			Name:       bucket.ObjectMeta.Name,
		}

		rsp, err := b.provisionerClient.DriverCreateBucket(ctx, req)
		if err != nil {
			if status.Code(err) != codes.AlreadyExists {
				klog.ErrorS(err, "Failed to create bucket",
					"bucket", bucket.ObjectMeta.Name)
				return errors.Wrap(err, "Failed to create bucket")
			}

		}

		if rsp == nil {
			err = errors.New("DriverCreateBucket returned a nil response")
			klog.ErrorS(err, "Internal Error")
			return err
		}

		if rsp.BucketId != "" {
			bucket.Status.BucketID = rsp.BucketId
			bucket.Status.BucketReady = true
		} else {
			err = errors.New("DriverCreateBucket returned an empty bucketID")
			return err
		}

		// Now we update the BucketReady status of BucketClaim
		if bucket.Spec.BucketClaim != nil {
			ref := bucket.Spec.BucketClaim
			bucketClaim, err := b.bucketClaims(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			bucketClaim.Status.BucketReady = true
			if _, err = b.bucketClaims(bucketClaim.Namespace).Update(ctx, bucketClaim, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	}

	controllerutil.AddFinalizer(bucket, consts.BucketFinalizer)
	if _, err = b.buckets().Update(ctx, bucket, metav1.UpdateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update bucket finalizers", "bucket", bucket.ObjectMeta.Name)
		return errors.Wrap(err, "Failed to update bucket finalizers")
	}

	// if this step fails, then controller will retry with backoff
	if _, err = b.buckets().UpdateStatus(ctx, bucket, metav1.UpdateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update bucket status",
			"bucket", bucket.ObjectMeta.Name)
		return errors.Wrap(err, "Failed to update bucket status")
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

	bucket := new.DeepCopy()

	var err error

	if !bucket.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(bucket, consts.BABucketFinalizer) {
			bucketClaimNs := bucket.Spec.BucketClaim.Namespace
			bucketClaimName := bucket.Spec.BucketClaim.Name
			bucketAccessList, err := b.bucketAccesses(bucketClaimNs).List(ctx, metav1.ListOptions{})

			for _, bucketAccess := range bucketAccessList.Items {
				if strings.EqualFold(bucketAccess.Spec.BucketClaimName, bucketClaimName) {
					err = b.bucketAccesses(bucketClaimNs).Delete(ctx, bucketAccess.Name, metav1.DeleteOptions{})
					if err != nil {
						return err
					}
				}
			}

			controllerutil.RemoveFinalizer(bucket, consts.BABucketFinalizer)
		}

		if controllerutil.ContainsFinalizer(bucket, consts.BucketFinalizer) {
			err = b.deleteBucketOp(ctx, bucket)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete attemps to delete a bucket. This function must be idempotent
// Return values
//    nil - Bucket successfully deleted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (b *BucketListener) Delete(ctx context.Context, inputBucket *v1alpha1.Bucket) error {
	klog.V(3).InfoS("Delete Bucket",
		"name", inputBucket.ObjectMeta.Name,
		"bucketclass", inputBucket.Spec.BucketClassName,
	)

	return nil

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

func (b *BucketListener) deleteBucketOp(ctx context.Context, bucket *v1alpha1.Bucket) error {
	if !strings.EqualFold(bucket.Spec.DriverName, b.driverName) {
		klog.V(5).InfoS("Skipping bucket for provisiner",
			"bucket", bucket.ObjectMeta.Name,
			"driver", bucket.Spec.DriverName,
		)
		return nil
	}

	// We ask the driver to clean up the bucket from the storage provider
	// only when the retain policy is set to Delete
	if bucket.Spec.DeletionPolicy == v1alpha1.DeletionPolicyDelete {
		req := &cosi.DriverDeleteBucketRequest{
			BucketId: bucket.Status.BucketID,
		}

		if _, err := b.provisionerClient.DriverDeleteBucket(ctx, req); err != nil {
			if status.Code(err) != codes.NotFound {
				klog.ErrorS(err, "Failed to delete bucket",
					"bucket", bucket.ObjectMeta.Name,
				)
				return err
			}
		}
	}

	if bucket.Spec.BucketClaim != nil {
		ref := bucket.Spec.BucketClaim
		bucketClaim, err := b.bucketClaims(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if controllerutil.RemoveFinalizer(bucketClaim, consts.BCFinalizer) {
			if _, err := b.bucketClaims(bucketClaim.ObjectMeta.Namespace).Update(ctx, bucketClaim, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *BucketListener) buckets() bucketapi.BucketInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().Buckets()
	}
	panic("uninitialized listener")
}

func (b *BucketListener) bucketClaims(namespace string) bucketapi.BucketClaimInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().BucketClaims(namespace)
	}

	panic("uninitialized listener")
}

func (b *BucketListener) bucketAccesses(namespace string) bucketapi.BucketAccessInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().BucketAccesses(namespace)
	}
	panic("uninitialized listener")
}
