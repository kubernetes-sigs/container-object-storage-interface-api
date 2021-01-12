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

package bucketaccess

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilversion "k8s.io/apimachinery/pkg/util/version"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-sigs/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1"
	bucketclientset "github.com/kubernetes-sigs/container-object-storage-interface-api/clientset"
	"github.com/kubernetes-sigs/container-object-storage-interface-api/controller"

	osspec "github.com/kubernetes-sigs/container-object-storage-interface-spec"

	"k8s.io/klog/v2"

	"golang.org/x/time/rate"
)

func generateSecretName(uid types.UID) string {
	return fmt.Sprintf("ba-%s", string(uid))
}

// bucketAccessListener manages BucketAccess objects
type bucketAccessListener struct {
	kubeClient         kubeclientset.Interface
	bucketAccessClient bucketclientset.Interface
	provisionerClient  osspec.ProvisionerClient

	// The name of the provisioner for which this controller handles
	// bucket access.
	provisionerName string
	kubeVersion     *utilversion.Version
}

// NewBucketAccessController returns a controller that manages BucketAccess objects
func NewBucketAccessController(provisionerName string, client osspec.ProvisionerClient) (*controller.ObjectStorageController, error) {
	rateLimit := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 60*time.Minute),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)

	identity := fmt.Sprintf("objectstorage-sidecar-%s", provisionerName)
	bc, err := controller.NewObjectStorageController(identity, "bucket-access-controller", 5, rateLimit)
	if err != nil {
		return nil, err
	}

	bal := bucketAccessListener{
		provisionerName:   provisionerName,
		provisionerClient: client,
	}
	bc.AddBucketAccessListener(&bal)

	return bc, nil
}

// InitializeKubeClient initializes the kubernetes client
func (bal *bucketAccessListener) InitializeKubeClient(k kubeclientset.Interface) {
	bal.kubeClient = k

	serverVersion, err := k.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("unable to get server version: %v", err)
	} else {
		bal.kubeVersion = utilversion.MustParseSemantic(serverVersion.GitVersion)
	}
}

// InitializeBucketClient initializes the object storage bucket client
func (bal *bucketAccessListener) InitializeBucketClient(bc bucketclientset.Interface) {
	bal.bucketAccessClient = bc
}

// Add will call the provisioner to grant permissions
func (bal *bucketAccessListener) Add(ctx context.Context, obj *v1alpha1.BucketAccess) error {
	klog.Infof("bucketAccessListener: add called for bucket access %s", obj.Name)

	// Verify this bucket access is for this provisioner
	if !strings.EqualFold(obj.Spec.Provisioner, bal.provisionerName) {
		return nil
	}

	bucketInstanceName := obj.Spec.BucketInstanceName
	bucket, err := bal.bucketAccessClient.ObjectstorageV1alpha1().Buckets().Get(ctx, bucketInstanceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get bucket instance %s: %+v", bucketInstanceName, err)
	}

	req := osspec.ProvisionerGrantBucketAccessRequest{
		Principal:     obj.Spec.Principal,
		AccessPolicy:  obj.Spec.PolicyActionsConfigMapData,
		BucketContext: obj.Spec.Parameters,
	}

	switch bucket.Spec.Protocol.Name {
	case v1alpha1.ProtocolNameS3:
		req.BucketName = bucket.Spec.Protocol.S3.BucketName
		req.BucketContext["Region"] = bucket.Spec.Protocol.S3.Region
		req.BucketContext["Version"] = bucket.Spec.Protocol.S3.Version
		req.BucketContext["SignatureVersion"] = string(bucket.Spec.Protocol.S3.SignatureVersion)
		req.BucketContext["Endpoint"] = bucket.Spec.Protocol.S3.Endpoint
	case v1alpha1.ProtocolNameAzure:
		req.BucketName = bucket.Spec.Protocol.AzureBlob.ContainerName
		req.BucketContext["StorageAccount"] = bucket.Spec.Protocol.AzureBlob.StorageAccount
	case v1alpha1.ProtocolNameGCS:
		req.BucketName = bucket.Spec.Protocol.GCS.BucketName
		req.BucketContext["ServiceAccount"] = bucket.Spec.Protocol.GCS.ServiceAccount
		req.BucketContext["PrivateKeyName"] = bucket.Spec.Protocol.GCS.PrivateKeyName
		req.BucketContext["ProjectID"] = bucket.Spec.Protocol.GCS.ProjectID
	default:
		return fmt.Errorf("unknown protocol: %s", bucket.Spec.Protocol.Name)
	}

	// TODO set grpc timeout
	rsp, err := bal.provisionerClient.ProvisionerGrantBucketAccess(ctx, &req)
	if err != nil {
		klog.Errorf("error calling ProvisionerGrantBucketAccess: %v", err)
		return err
	}
	klog.Infof("provisioner returned grant bucket access response %v", rsp)

	// Only update the principal in the BucketAccess if it wasn't set because
	// that means that the provisioner created one
	if len(obj.Spec.Principal) == 0 && obj.Spec.ServiceAccount == nil {
		err = bal.updatePrincipal(ctx, obj.Name, rsp)
		if err != nil {
			return err
		}
	}

	// Only create the secret with credentials if serviveAccount isn't set.
	// If serviceAccount is set then authorization happens out of band in the
	// cloud provider
	if obj.Spec.ServiceAccount == nil {
		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: generateSecretName(obj.UID),
			},
			StringData: map[string]string{
				"CredentialsFilePath":     rsp.CredentialsFilePath,
				"CredentialsFileContents": rsp.CredentialsFileContents,
			},
			Type: corev1.SecretTypeOpaque,
		}

		// It's unlikely but should probably handle retries on rare case of collision
		_, err = bal.kubeClient.CoreV1().Secrets("objectstorage-system").Create(ctx, &secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		// TODO update the mintedSecretName in the BA
	}

	// update bucket access status to granted
	return bal.updateStatus(ctx, obj.Name, "Permissions Granted", true)
}

// Update does nothing
func (bal *bucketAccessListener) Update(ctx context.Context, old, new *v1alpha1.BucketAccess) error {
	klog.Infof("bucketAccessListener: update called for bucket %s", old.Name)
	return nil
}

// Delete will call the provisioner to revoke permissions
func (bal *bucketAccessListener) Delete(ctx context.Context, obj *v1alpha1.BucketAccess) error {
	klog.Infof("bucketAccessListener: delete called for bucket access %s", obj.Name)

	// Verify this bucket access is for this provisioner
	if !strings.EqualFold(obj.Spec.Provisioner, bal.provisionerName) {
		return nil
	}

	bucketInstanceName := obj.Spec.BucketInstanceName
	bucket, err := bal.bucketAccessClient.ObjectstorageV1alpha1().Buckets().Get(ctx, bucketInstanceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get bucket instance %s: %+v", bucketInstanceName, err)
	}

	req := osspec.ProvisionerRevokeBucketAccessRequest{
		Principal:     obj.Spec.Principal,
		BucketContext: obj.Spec.Parameters,
	}

	switch bucket.Spec.Protocol.Name {
	case v1alpha1.ProtocolNameS3:
		req.BucketName = bucket.Spec.Protocol.S3.BucketName
		req.BucketContext["Region"] = bucket.Spec.Protocol.S3.Region
		req.BucketContext["Version"] = bucket.Spec.Protocol.S3.Version
		req.BucketContext["SignatureVersion"] = string(bucket.Spec.Protocol.S3.SignatureVersion)
		req.BucketContext["Endpoint"] = bucket.Spec.Protocol.S3.Endpoint
	case v1alpha1.ProtocolNameAzure:
		req.BucketName = bucket.Spec.Protocol.AzureBlob.ContainerName
		req.BucketContext["StorageAccount"] = bucket.Spec.Protocol.AzureBlob.StorageAccount
	case v1alpha1.ProtocolNameGCS:
		req.BucketName = bucket.Spec.Protocol.GCS.BucketName
		req.BucketContext["ServiceAccount"] = bucket.Spec.Protocol.GCS.ServiceAccount
		req.BucketContext["PrivateKeyName"] = bucket.Spec.Protocol.GCS.PrivateKeyName
		req.BucketContext["ProjectID"] = bucket.Spec.Protocol.GCS.ProjectID
	default:
		return fmt.Errorf("unknown protocol: %s", bucket.Spec.Protocol.Name)
	}

	// TODO set grpc timeout
	rsp, err := bal.provisionerClient.ProvisionerRevokeBucketAccess(ctx, &req)
	if err != nil {
		klog.Errorf("error calling ProvisionerRevokeBucketAccess: %v", err)
		return err
	}
	klog.Infof("provisioner returned revoke bucket access response %v", rsp)

	// Delete the secret
	if obj.Spec.ServiceAccount == nil {
		// TODO get the minted secret name from the BA

		// It's unlikely but should probably handle retries on rare case of collision
		err = bal.kubeClient.CoreV1().Secrets("objectstorage-system").Delete(ctx, generateSecretName(obj.UID), metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	// Update bucket access status to revoked
	return bal.updateStatus(ctx, obj.Name, "Permissions Revoked", false)
}

func (bal *bucketAccessListener) updateStatus(ctx context.Context, name, msg string, state bool) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		bucketAccess, err := bal.bucketAccessClient.ObjectstorageV1alpha1().BucketAccesses().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		bucketAccess.Status.Message = msg
		bucketAccess.Status.AccessGranted = state
		_, err = bal.bucketAccessClient.ObjectstorageV1alpha1().BucketAccesses().UpdateStatus(ctx, bucketAccess, metav1.UpdateOptions{})
		return err
	})
	return err
}

func (bal *bucketAccessListener) updatePrincipal(ctx context.Context, name string, resp *osspec.ProvisionerGrantBucketAccessResponse) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		bucketAccess, err := bal.bucketAccessClient.ObjectstorageV1alpha1().BucketAccesses().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		bucketAccess.Spec.Principal = resp.Principal
		_, err = bal.bucketAccessClient.ObjectstorageV1alpha1().BucketAccesses().Update(ctx, bucketAccess, metav1.UpdateOptions{})
		return err
	})
	return err
}
