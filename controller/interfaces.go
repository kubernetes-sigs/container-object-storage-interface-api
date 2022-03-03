package controller

import (
	"context"

	// storage
	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1"
	bucketclientset "sigs.k8s.io/container-object-storage-interface-api/clientset"

	// k8s client
	kubeclientset "k8s.io/client-go/kubernetes"
)

// Set the clients for each of the listeners
type GenericListener interface {
	InitializeKubeClient(kubeclientset.Interface)
	InitializeBucketClient(bucketclientset.Interface)
}

type BucketListener interface {
	GenericListener

	Add(ctx context.Context, b *v1alpha1.Bucket) error
	Update(ctx context.Context, old *v1alpha1.Bucket, new *v1alpha1.Bucket) error
	Delete(ctx context.Context, b *v1alpha1.Bucket) error
}

func (c *ObjectStorageController) AddBucketListener(b BucketListener) {
	c.initialized = true
	c.BucketListener = b
}

type BucketClaimListener interface {
	GenericListener

	Add(ctx context.Context, b *v1alpha1.BucketClaim) error
	Update(ctx context.Context, old *v1alpha1.BucketClaim, new *v1alpha1.BucketClaim) error
	Delete(ctx context.Context, b *v1alpha1.BucketClaim) error
}

func (c *ObjectStorageController) AddBucketClaimListener(b BucketClaimListener) {
	c.initialized = true
	c.BucketClaimListener = b
}

type BucketAccessListener interface {
	GenericListener

	Add(ctx context.Context, b *v1alpha1.BucketAccess) error
	Update(ctx context.Context, old *v1alpha1.BucketAccess, new *v1alpha1.BucketAccess) error
	Delete(ctx context.Context, b *v1alpha1.BucketAccess) error
}

func (c *ObjectStorageController) AddBucketAccessListener(b BucketAccessListener) {
	c.initialized = true
	c.BucketAccessListener = b
}
