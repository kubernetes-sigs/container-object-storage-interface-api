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

package bucketaccess

import (
	"context"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	kube "k8s.io/client-go/kubernetes"
	kubecorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1"
	buckets "sigs.k8s.io/container-object-storage-interface-api/clientset"
	bucketapi "sigs.k8s.io/container-object-storage-interface-api/clientset/typed/objectstorage.k8s.io/v1alpha1"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CredentialsFilePath     = "CredentialsFilePath"
	CredentialsFileContents = "CredentialsFileContents"
)

// BucketAccessListener manages Bucket objects
type BucketAccessListener struct {
	provisionerClient cosi.ProvisionerClient
	provisionerName   string

	kubeClient   kube.Interface
	bucketClient buckets.Interface
	kubeVersion  *utilversion.Version

	namespace string
}

// NewBucketAccessListener returns a resource handler for BucketAccess objects
func NewBucketAccessListener(provisionerName string, client cosi.ProvisionerClient) (*BucketAccessListener, error) {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return nil, errors.New("POD_NAMESPACE env var cannot be empty")
	}

	return &BucketAccessListener{
		provisionerName:   provisionerName,
		provisionerClient: client,
		namespace:         ns,
	}, nil
}

// Add attempts to provision credentials to access a given bucket. This function must be idempotent
// Return values
//    nil - BucketAccess successfully granted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Add(ctx context.Context, inputBucketAccess *v1alpha1.BucketAccess) error {
	bucketAccess := inputBucketAccess.DeepCopy()

	bucketName := bucketAccess.Spec.BucketName
	klog.V(3).InfoS("Add BucketAccess",
		"name", bucketAccess.Name,
		"bucket", bucketName,
	)

	if bucketAccess.Spec.MintedSecretName != "" {
		klog.V(5).InfoS("AccessAlreadyGranted",
			"bucketAccess", bucketAccess.Name,
			"bucket", bucketName,
		)
		return nil
	}

	bucket, err := bal.Buckets().Get(ctx, bucketName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to fetch bucket", "bucket", bucketName)
		return errors.Wrap(err, "Failed to fetch bucket")
	}

	if !strings.EqualFold(bucket.Spec.Provisioner, bal.provisionerName) {
		klog.V(5).InfoS("Skipping bucketaccess for provisiner",
			"bucketAccess", bucketAccess.Name,
			"provisioner", bucket.Spec.Provisioner,
		)
		return nil
	}

	if bucketAccess.Status.AccessGranted {
		klog.V(5).InfoS("AccessAlreadyGranted",
			"bucketaccess", bucketAccess.Name,
			"bucket", bucket.Name,
		)
		return nil
	}

	if bucket.Spec.BucketID == "" {
		err := errors.New("BucketID cannot be empty")
		klog.ErrorS(err,
			"Invalid arguments",
			"bucket", bucket.Name,
			"bucketAccess", bucketAccess.Name,
		)
		return errors.Wrap(err, "Invalid arguments")
	}

	req := &cosi.ProvisionerGrantBucketAccessRequest{
		BucketId:     bucket.Spec.BucketID,
		AccountName:  bucketAccess.Name,
		AccessPolicy: bucketAccess.Spec.PolicyActionsConfigMapData,
	}

	// This needs to be idempotent
	rsp, err := bal.provisionerClient.ProvisionerGrantBucketAccess(ctx, req)
	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			klog.ErrorS(err,
				"Failed to grant access",
				"bucketAccess", bucketAccess.Name,
				"bucket", bucket.Name,
			)
			return errors.Wrap(err, "failed to grant access")
		}

	}
	ns := bal.namespace
	mintedSecretName := "ba-" + string(bucketAccess.UID)
	if _, err := bal.Secrets(ns).Get(ctx, mintedSecretName, metav1.GetOptions{}); err != nil {
		if !kubeerrors.IsNotFound(err) {
			klog.ErrorS(err,
				"Failed to create secrets",
				"bucketAccess", bucketAccess.Name,
				"bucket", bucket.Name)
			return errors.Wrap(err, "failed to fetch secrets")
		}

		// if secret doesn't exist, create it
		credentialsFileContents := rsp.CredentialsFileContents
		credentialsFilePath := rsp.CredentialsFilePath

		if _, err := bal.Secrets(ns).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mintedSecretName,
				Namespace: ns,
			},
			StringData: map[string]string{
				CredentialsFilePath:     credentialsFilePath,
				CredentialsFileContents: credentialsFileContents,
			},
			Type: corev1.SecretTypeOpaque,
		}, metav1.CreateOptions{}); err != nil {
			if !kubeerrors.IsAlreadyExists(err) {
				klog.ErrorS(err,
					"Failed to create minted secret",
					"bucketAccess", bucketAccess.Name,
					"bucket", bucket.Name)
				return errors.Wrap(err, "Failed to create minted secret")
			}
		}
	}

	bucketAccess.Spec.AccountID = rsp.AccountId
	bucketAccess.Status.AccessGranted = true
	bucketAccess.Spec.MintedSecretName = mintedSecretName

	// if this step fails, then controller will retry with backoff
	if _, err := bal.BucketAccesses().Update(ctx, bucketAccess, metav1.UpdateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update BucketAccess",
			"bucketAccess", bucketAccess.Name,
			"bucket", bucket.Name)
		return errors.Wrap(err, "Failed to update BucketAccess")
	}

	return nil
}

// Update attempts to reconcile changes to a given bucketAccess. This function must be idempotent
// Return values
//    nil - BucketAccess successfully reconciled
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Update(ctx context.Context, old, new *v1alpha1.BucketAccess) error {
	klog.V(3).InfoS("Update BucketAccess",
		"name", old.Name)

	return nil
}

// Delete attemps to delete a bucketAccess. This function must be idempotent
// Return values
//    nil - BucketAccess successfully deleted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Delete(ctx context.Context, bucketAccess *v1alpha1.BucketAccess) error {
	klog.V(3).InfoS("Delete BucketAccess",
		"name", bucketAccess.Name,
		"bucket", bucketAccess.Spec.BucketName,
	)
	return nil
}

func (b *BucketAccessListener) Secrets(ns string) kubecorev1.SecretInterface {
	if b.kubeClient != nil {
		return b.kubeClient.CoreV1().Secrets(ns)
	}
	panic("uninitialized listener")
}

func (b *BucketAccessListener) BucketAccesses() bucketapi.BucketAccessInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().BucketAccesses()
	}
	panic("uninitialized listener")
}

func (b *BucketAccessListener) Buckets() bucketapi.BucketInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().Buckets()
	}
	panic("uninitialized listener")
}

// InitializeKubeClient initializes the kubernetes client
func (b *BucketAccessListener) InitializeKubeClient(k kube.Interface) {
	b.kubeClient = k

	serverVersion, err := k.Discovery().ServerVersion()
	if err != nil {
		klog.ErrorS(err, "Cannot determine server version")
	} else {
		b.kubeVersion = utilversion.MustParseSemantic(serverVersion.GitVersion)
	}
}

// InitializeBucketClient initializes the object storage bucket client
func (b *BucketAccessListener) InitializeBucketClient(bc buckets.Interface) {
	b.bucketClient = bc
}
