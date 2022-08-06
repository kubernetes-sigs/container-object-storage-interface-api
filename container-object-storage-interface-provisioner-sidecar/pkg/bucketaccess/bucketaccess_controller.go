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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	kube "k8s.io/client-go/kubernetes"
	kubecorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"

	cosiapi "sigs.k8s.io/container-object-storage-interface-api/apis"
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

// BucketAccessListener manages Bucket objects
type BucketAccessListener struct {
	provisionerClient cosi.ProvisionerClient
	driverName        string

	kubeClient   kube.Interface
	bucketClient buckets.Interface
	kubeVersion  *utilversion.Version
}

// NewBucketAccessListener returns a resource handler for BucketAccess objects
func NewBucketAccessListener(driverName string, client cosi.ProvisionerClient) (*BucketAccessListener, error) {
	return &BucketAccessListener{
		driverName:        driverName,
		provisionerClient: client,
	}, nil
}

// Add attempts to provision credentials to access a given bucket. This function must be idempotent
// Return values
//    nil - BucketAccess successfully granted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Add(ctx context.Context, inputBucketAccess *v1alpha1.BucketAccess) error {
	bucketAccess := inputBucketAccess.DeepCopy()

	if bucketAccess.Status.AccessGranted && bucketAccess.Status.AccountID != "" {
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
	if secretCredName == "" {
		return errors.New("CredentialsSecretName not defined in the BucketAccess")
	}

	bucketAccessClass, err := bal.bucketAccessClasses().Get(ctx, bucketAccessClassName, metav1.GetOptions{})
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

	namespace := bucketAccess.ObjectMeta.Namespace
	bucketClaim, err := bal.bucketClaims(namespace).Get(ctx, bucketClaimName, metav1.GetOptions{})
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

	authType := cosi.AuthenticationType_UnknownAuthenticationType
	if bucketAccessClass.AuthenticationType == v1alpha1.AuthenticationTypeKey {
		authType = cosi.AuthenticationType_Key
	} else if bucketAccessClass.AuthenticationType == v1alpha1.AuthenticationTypeIAM {
		authType = cosi.AuthenticationType_IAM
	}

	if authType == cosi.AuthenticationType_IAM && bucketAccess.Spec.ServiceAccountName == "" {
		return errors.New("Must define ServiceAccountName when AuthenticationType is IAM")
	}

	if bucketAccess.Status.AccessGranted == true {
		klog.V(5).InfoS("AccessAlreadyGranted",
			"bucketAccess", bucketAccess.ObjectMeta.Name,
			"bucketClaim", bucketClaimName,
		)
		return nil
	}

	bucket, err := bal.buckets().Get(ctx, bucketClaim.Status.BucketName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to fetch bucket", "bucket", bucketClaim.Status.BucketName)
		return errors.Wrap(err, "Failed to fetch bucket")
	}

	if bucket.Status.BucketReady != true || bucket.Status.BucketID == "" {
		return errors.New("BucketAccess can't be granted to bucket not in Ready state and without a bucketID")
	}

	accountName := consts.AccountNamePrefix + string(bucketAccess.UID)

	req := &cosi.DriverGrantBucketAccessRequest{
		BucketId:           bucket.Status.BucketID,
		Name:               accountName,
		AuthenticationType: authType,
		Parameters:         bucketAccessClass.Parameters,
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

	if rsp.AccountId == "" {
		err = errors.New("AccountId not defined in DriverGrantBucketAccess")
		klog.ErrorS(err, "BucketAccess", bucketAccess.ObjectMeta.Name)
		return errors.Wrap(err, fmt.Sprintf("BucketAccess %s", bucketAccess.ObjectMeta.Name))
	}

	credentials := rsp.Credentials
	if len(credentials) != 1 {
		err = errors.New("Credentials returned in DriverGrantBucketAccessResponse should be of length 1")
		klog.ErrorS(err, "BucketAccess", bucketAccess.ObjectMeta.Name)
		return errors.Wrap(err, fmt.Sprintf("BucketAccess %s", bucketAccess.ObjectMeta.Name))
	}

	bucketInfoName := consts.BucketInfoPrefix + string(bucketAccess.ObjectMeta.UID)

	bucketInfo := cosiapi.BucketInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name: bucketInfoName,
		},
		Spec: cosiapi.BucketInfoSpec{
			BucketName:         bucket.ObjectMeta.Name,
			AuthenticationType: bucketAccessClass.AuthenticationType,
			Protocols:          []v1alpha1.Protocol{bucketAccess.Spec.Protocol},
		},
	}

	var val *cosi.CredentialDetails
	var ok bool

	if val, ok = credentials[consts.S3Key]; ok {
		secretS3 := &cosiapi.SecretS3{
			Endpoint:        "https://s3.amazonaws.com",
			Region:          "us-west-1",
			AccessKeyID:     val.Secrets[consts.S3SecretAccessKeyID],
			AccessSecretKey: val.Secrets[consts.S3SecretAccessSecretKey],
		}

		bucketInfo.Spec.S3 = secretS3
	} else if val, ok = credentials[consts.AzureKey]; ok {
		expiryTs := val.Secrets[consts.AzureSecretExpiryTimeStamp]
		expiryTimestamp, _ := time.Parse(consts.DefaultTimeFormat, expiryTs)
		metav1Time := &metav1.Time{Time: expiryTimestamp}
		secretAzure := &cosiapi.SecretAzure{
			AccessToken:     val.Secrets[consts.AzureSecretAccessToken],
			ExpiryTimeStamp: metav1Time,
		}

		bucketInfo.Spec.Azure = secretAzure
	}

	stringData, err := json.Marshal(bucketInfo)
	if err != nil {
		return errors.New("Error converting bucketinfo into secret")
	}

	if _, err := bal.secrets(namespace).Get(ctx, secretCredName, metav1.GetOptions{}); err != nil {
		if !kubeerrors.IsNotFound(err) {
			klog.ErrorS(err,
				"Failed to create secrets",
				"bucketAccess", bucketAccess.ObjectMeta.Name,
				"bucket", bucket.ObjectMeta.Name)
			return errors.Wrap(err, "failed to fetch secrets")
		}

		if _, err := bal.secrets(namespace).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:       secretCredName,
				Namespace:  namespace,
				Finalizers: []string{consts.SecretFinalizer},
			},
			StringData: map[string]string{
				"BucketInfo": string(stringData),
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

	if controllerutil.AddFinalizer(bucket, consts.BABucketFinalizer) {
		_, err = bal.buckets().Update(ctx, bucket, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	if controllerutil.AddFinalizer(bucketAccess, consts.BAFinalizer) {
		if _, err = bal.bucketAccesses(bucketAccess.ObjectMeta.Namespace).Update(ctx, bucketAccess, metav1.UpdateOptions{}); err != nil {
			klog.ErrorS(err, "Failed to update BucketAccess finalizer",
				"bucketAccess", bucketAccess.ObjectMeta.Name,
				"bucket", bucket.ObjectMeta.Name)
			return errors.Wrap(err, fmt.Sprintf("Failed to update BucketAccess finalizer. BucketAccess: %s", bucketAccess.ObjectMeta.Name))
		}
	}

	bucketAccess.Status.AccountID = rsp.AccountId
	bucketAccess.Status.AccessGranted = true

	// if this step fails, then controller will retry with backoff
	if _, err := bal.bucketAccesses(bucketAccess.ObjectMeta.Namespace).UpdateStatus(ctx, bucketAccess, metav1.UpdateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to update BucketAccess Status",
			"bucketAccess", bucketAccess.ObjectMeta.Name,
			"bucket", bucket.ObjectMeta.Name)
		return errors.Wrap(err, fmt.Sprintf("Failed to update BucketAccess Status. BucketAccess: %s", bucketAccess.ObjectMeta.Name))
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

	bucketAccess := new.DeepCopy()
	err := bal.deleteBucketAccessOp(ctx, bucketAccess)
	if err != nil {
		return err
	}

	return nil
}

// Delete attemps to delete a bucketAccess. This function must be idempotent
// Return values
//    nil - BucketAccess successfully deleted
//    non-nil err - Internal error                                [requeue'd with exponential backoff]
func (bal *BucketAccessListener) Delete(ctx context.Context, bucketAccess *v1alpha1.BucketAccess) error {
	klog.V(3).InfoS("Delete BucketAccess",
		"name", bucketAccess.ObjectMeta.Name,
		"bucketClaim", bucketAccess.Spec.BucketClaimName,
	)

	return nil
}

func (bal *BucketAccessListener) deleteBucketAccessOp(ctx context.Context, bucketAccess *v1alpha1.BucketAccess) error {
	credSecretName := bucketAccess.Spec.CredentialsSecretName
	secret, err := bal.secrets(bucketAccess.ObjectMeta.Namespace).Get(ctx, credSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if controllerutil.RemoveFinalizer(secret, consts.SecretFinalizer) {
		_, err = bal.secrets(bucketAccess.ObjectMeta.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	if controllerutil.RemoveFinalizer(bucketAccess, consts.BAFinalizer) {
		_, err = bal.bucketAccesses(bucketAccess.ObjectMeta.Namespace).Update(ctx, bucketAccess, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (bal *BucketAccessListener) secrets(ns string) kubecorev1.SecretInterface {
	if bal.kubeClient != nil {
		return bal.kubeClient.CoreV1().Secrets(ns)
	}
	panic("uninitialized listener")
}

func (bal *BucketAccessListener) bucketAccesses(ns string) bucketapi.BucketAccessInterface {
	if bal.bucketClient != nil {
		return bal.bucketClient.ObjectstorageV1alpha1().BucketAccesses(ns)
	}
	panic("uninitialized listener")
}

func (bal *BucketAccessListener) buckets() bucketapi.BucketInterface {
	if bal.bucketClient != nil {
		return bal.bucketClient.ObjectstorageV1alpha1().Buckets()
	}
	panic("uninitialized listener")
}

func (bal *BucketAccessListener) bucketClaims(namespace string) bucketapi.BucketClaimInterface {
	if bal.bucketClient != nil {
		return bal.bucketClient.ObjectstorageV1alpha1().BucketClaims(namespace)
	}
	panic("uninitialized listener")
}

func (bal *BucketAccessListener) bucketAccessClasses() bucketapi.BucketAccessClassInterface {
	if bal.bucketClient != nil {
		return bal.bucketClient.ObjectstorageV1alpha1().BucketAccessClasses()
	}
	panic("uninitialized listener")
}

// InitializeKubeClient initializes the kubernetes client
func (bal *BucketAccessListener) InitializeKubeClient(k kube.Interface) {
	bal.kubeClient = k

	serverVersion, err := k.Discovery().ServerVersion()
	if err != nil {
		klog.ErrorS(err, "Cannot determine server version")
	} else {
		bal.kubeVersion = utilversion.MustParseSemantic(serverVersion.GitVersion)
	}
}

// InitializeBucketClient initializes the object storage bucket client
func (bal *BucketAccessListener) InitializeBucketClient(bc buckets.Interface) {
	bal.bucketClient = bc
}
