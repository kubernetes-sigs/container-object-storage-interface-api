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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	kube "k8s.io/client-go/kubernetes"
	kubecorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"

	cosiapi "sigs.k8s.io/container-object-storage-interface-api/apis"
	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1"
	buckets "sigs.k8s.io/container-object-storage-interface-api/clientset"
	bucketapi "sigs.k8s.io/container-object-storage-interface-api/clientset/typed/objectstorage.k8s.io/v1alpha1"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	accountNamePrefix = "ba-"
	baFinalizer = "cosi.objectstorage.k8s.io/bucketaccess-protection-"
	secretFinalizer = "cosi.objectstorage.k8s.io/secret-protection"
)

// BucketAccessListener manages Bucket objects
type BucketAccessListener struct {
	provisionerClient cosi.ProvisionerClient
	driverName   string

	kubeClient   kube.Interface
	bucketClient buckets.Interface
	kubeVersion  *utilversion.Version
}

// NewBucketAccessListener returns a resource handler for BucketAccess objects
func NewBucketAccessListener(driverName string, client cosi.ProvisionerClient) (*BucketAccessListener, error) {
	return &BucketAccessListener{
		driverName:   driverName,
		provisionerClient: client,
	}, nil
}

// Add attempts to provision credentials to access a given bucket. This function must be idempotent
// Return values
//    nil - BucketAccess successfully granted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Add(ctx context.Context, inputBucketAccess *v1alpha1.BucketAccess) error {
	bucketAccess := inputBucketAccess.DeepCopy()

	if bucketAccess.Status.AccessGranted && bucketAccess.Status.AccountID != nil {
		klog.V(3).InfoS("BucketAccess already exists", bucketAccess.ObjectMeta.Name)
		return nil
	}

	bucketClaimName := bucketAccess.Spec.BucketClaimName
	klog.V(3).InfoS("Add BucketAccess",
		"name", bucketAccess.ObjectMeta.Name,
		"bucketClaim", bucketClaimName,
	)

	bucketAccessClassName := bucketAccess.Spec.BucketAccessClassName
	klog.V(3).InfoS("Add BucketAccess",
		"name", bucketAccess.ObjectMeta.Name,
		"BucketAccessClassName", bucketAccessClassName,
	)

	secretCredName := bucketAccess.Spec.CredentialsSecretName
	if secretCredName == nil {
		return errors.New("CredentialsSecretName not defined in the BucketAccess")
	}

	authType := cosi.AuthenticationType_UnknownAuthenticationType
	if bucketAccess.Spec.AuthenticationType == v1alpha1.AuthenticationTypeKey {
		authType = cosi.AuthenticationType_Key
	} else if bucketAccess.Spec.AuthenticationType == v1alpha1.AuthenticationTypeIAM {
		authType = cosi.AuthenticationType_IAM
	}

	if authType == cosi.AuthenticationType_IAM && bucketAccess.Spec.ServiceAccountName == "" {
		return errors.New("Must define ServiceAccountName when AuthenticationType is IAM")
	}

	namespace := bucketAccess.ObjectMeta.Namespace
	bucketClaim, err := bal.BucketClaims(namespace).Get(ctx, bucketClaimName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to fetch bucketClaim", "bucketClaim", bucketClaimName)
		return errors.Wrap(err, "Failed to fetch bucketClaim")
	}


	if bucketClaim.Status.BucketName == "" || bucketClaim.Status.BucketReady != true {
		err := errors.New("BucketName cannot be empty or BucketNotReady in bucketClaim")
		klog.ErrorS(err,
			"Invalid arguments",
			"bucketClaim", bucketClaim.Name,
			"bucketAccess", bucketAccess.ObjectMeta.Name,
		)
		return errors.Wrap(err, "Invalid arguments")
	}

	bucketAccessClass, err := bal.BucketAccessClasses().Get(ctx, bucketAccessClassName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to fetch bucketAccessClass", "bucketAccessClass", bucketAccessClassName)
		return errors.Wrap(err, "Failed to fetch BucketAccessClass")
	}

	if !strings.EqualFold(bucketAccessClass.DriverName, bal.driverName) {
		klog.V(5).InfoS("Skipping bucketaccess for driver",
			"bucketAccess", bucketAccess.ObjectMeta.Name,
			"driver", bucketAccessClass.DriverName,
		)
		return nil
	}


	if bucketAccess.Status.AccessGranted == true {
		klog.V(5).InfoS("AccessAlreadyGranted",
			"bucketAccess", bucketAccess.ObjectMeta.Name,
			"bucketClaim", bucketClaimName,
		)
		return nil
	}

	bucket, err := bal.Buckets().Get(ctx, bucketClaim.Status.BucketName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to fetch bucket", "bucket", bucketClaim.Status.BucketName)
		return errors.Wrap(err, "Failed to fetch bucket")
	}

	if bucket.Status.BucketStatus != true || bucket.Status.BucketID == "" {
		return errors.New("BucketAccess can't be granted to bucket not in Ready state and without a bucketID")
	}

	accountName := accountNamePrefix + string(bucketAccess.UID)

	req := &cosi.DriverGrantBucketAccessRequest{
		BucketId:     bucket.Status.BucketID,
		AccountName:  accountName,
		AuthenticationType: authType,
		Parameters: bucketAccessClass.Parameters,
	}

	// This needs to be idempotent
	rsp, err := bal.provisionerClient.DriverGrantBucketAccess(ctx, req)
	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			klog.ErrorS(err,
				"Failed to grant access",
				"bucketAccess", bucketAccess.ObjectMeta.Name,
				"bucketClaim", bucketClaimName,
			)
			return errors.Wrap(err, "failed to grant access")
		}

	}

	if rsp.AccountId == nil {
		klog.ErrorS("AccountId not defined in DriverGrantBucketAccess")
		return errors.New("Failed to grant access. AccountId not defined in DriverGrantBucketAccess.")
	}

	bucketInfo := cosiapi.BucketInfo {
		ObjectMeta: metav1.ObjectMeta {
			name: secretCredName,
		},
		BucketInfoSpec: cosiapi.BucketInfoSpec {
			BucketName: bucket.ObjectMeta.Name,
			AuthenticationType: bucketAccess.Spec.AuthenticationType,
			Endpoint: ,
			Region: ,
			Protocol: ,
		}
	}

	srtingData, err := json.Marshal(bucketInfo)
	if err != nil {
		return errors.New("Error converting bucketinfo into secret")
	}

	if _, err := bal.Secrets(namespace).Get(ctx, secretCredName, metav1.GetOptions{}); err != nil {
		if !kubeerrors.IsNotFound(err) {
			klog.ErrorS(err,
				"Failed to create secrets",
				"bucketAccess", bucketAccess.ObjectMeta.Name,
				"bucket", bucket.ObjectMeta.Name)
			return errors.Wrap(err, "failed to fetch secrets")
		}

		if _, err := bal.Secrets(namespace).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretCredName,
				Namespace: namespace,
				Finalizers: []string{secretFinalizer},
			},
			StringData: map[string]string{
				BucketInfo: string(stringData),
			},
			Type: corev1.SecretTypeOpaque,
		}, metav1.CreateOptions{}); err != nil {
			if !kubeerrors.IsAlreadyExists(err) {
				klog.ErrorS(err,
					"Failed to create minted secret",
					"bucketAccess", bucketAccess.ObjectMeta.Name,
					"bucket", bucket.ObjectMeta.Name)
				return errors.Wrap(err, "Failed to create minted secret")
			}
		}
	}

	bucketFinalizer := baFinalizer + string(bucketAccess.ObjectMeta.UID)
	finalizers := bucket.ObjectMeta.Finalizers
	finalizers = append(finalizers, bucketFinalizer)
	bucket.ObjectMeta.Finalizers = finalizers
	_, err = bal.Buckets().Update(ctx, bucket, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	bucketAccess.Status.AccountID = rsp.AccountId
	bucketAccess.Status.AccessGranted = true

	// if this step fails, then controller will retry with backoff
	if _, err := bal.BucketAccesses().UpdateStatus(ctx, bucketAccess, metav1.UpdateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update BucketAccess Status",
			"bucketAccess", bucketAccess.ObjectMeta.Name,
			"bucket", bucket.ObjectMeta.Name)
		return errors.Wrap(err, "Failed to update BucketAccess Status")
	}

	return nil
}

// Update attempts to reconcile changes to a given bucketAccess. This function must be idempotent
// Return values
//    nil - BucketAccess successfully reconciled
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Update(ctx context.Context, old, new *v1alpha1.BucketAccess) error {
	klog.V(3).InfoS("Update BucketAccess",
		"name", old.ObjectMeta.Name)

	return nil
}

// Delete attemps to delete a bucketAccess. This function must be idempotent
// Return values
//    nil - BucketAccess successfully deleted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Delete(ctx context.Context, bucketAccess *v1alpha1.BucketAccess) error {
	klog.V(3).InfoS("Delete BucketAccess",
		"name", bucketAccess.ObjectMeta.Name,
		"bucket", bucketAccess.Spec.BucketName,
	)

	credSecretName := bucketAccess.Spec.CredentialsSecretName
	secret, err := bal.Secrets(bucketAccess.ObjectMeta.Namespace).Get(ctx, credSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if controllerutil.RemoveFinalizer(secret, secretFinalizer) {
		_, err = bal.Secrets(bucketAccess.ObjectMeta.Namespace).Update(ctx, credSecretName, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	bucketClaimName = bucketAccess.Spec.BucketClaimName
	bucketClaim, err := bal.BucketClaims(bucketAccess.ObjectMeta.Namespace).Get(ctx, bucketClaimName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	bucket, err := bal.Buckets().Get(ctx, bucketClaim.Status.BucketName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	bucketFinalizer := baFinalizer + string(bucketAccess.ObjectMeta.UID)
	if controllerutil.RemoveFinalizer(bucketFinalizer) {
		_, err = bal.Buckets().Update(ctx, bucket, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

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

func (b *BucketAccessListener) BucketClaims(namespace string) bucketapi.BucketClaimInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().BucketClaims(namespace)
	}
	panic("uninitialized listener")
}

func (b *BucketAccessListener) BucketAccessClasses() bucketapi.BucketClaimInterface {
	if b.bucketClient != nil {
		return b.bucketClient.ObjectstorageV1alpha1().BucketAccessClasses()
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
