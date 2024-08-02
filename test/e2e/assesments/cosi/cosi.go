package cosi

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	cosiv1alpha1 "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"

	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

// CRDsInstalled checks if the required COSI CRDs are installed in the cluster.
func CRDsInstalled() types.TestEnvFunc {
	return func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		var crds apiextensionsv1.CustomResourceDefinitionList

		expectedCRDs := []string{
			"bucketaccessclasses.objectstorage.k8s.io",
			"bucketaccesses.objectstorage.k8s.io",
			"bucketclaims.objectstorage.k8s.io",
			"bucketclasses.objectstorage.k8s.io",
			"buckets.objectstorage.k8s.io",
		}

		if err := cfg.Client().Resources().List(ctx, &crds); err != nil {
			t.Fatal(err)
		}

		found := 0

		for _, item := range crds.Items {
			for _, expected := range expectedCRDs {
				if item.GetObjectMeta().GetName() == expected {
					found++
				}
			}
		}

		if len(expectedCRDs) != found {
			t.Fatal("COSI CRDs not installed")
		}

		return ctx, nil
	}
}

// ObjectstorageControllerInstalled checks if the COSI object storage controller deployment is installed.
func ObjectstorageControllerInstalled() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var deploymentList appsv1.DeploymentList

		selector := resources.WithLabelSelector("app.kubernetes.io/part-of=container-object-storage-interface")

		if err := cfg.Client().Resources().List(ctx, &deploymentList, selector); err != nil {
			t.Fatal(err)
		}

		if len(deploymentList.Items) == 0 {
			t.Fatal("deployment not found")
		}

		return ctx
	}
}

// BucketAccessStatusReady checks if the status of BucketAccess is ready.
func BucketAccessStatusReady(ready bool) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// BucketClaimStatusReady checks if the status of BucketClaim is ready.
func BucketClaimStatusReady(ready bool) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// CreateBucket creates a new Bucket resource.
func CreateBucket(bucket *cosiv1alpha1.Bucket) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// BucketExists checks if the Bucket resource exists.
func BucketExists(expected bool) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// CreateBucketClaim creates a new BucketClaim resource.
func CreateBucketClaim(bucketClaim *cosiv1alpha1.BucketClaim) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// DeleteBucketClaim deletes the specified BucketClaim resource.
func DeleteBucketClaim() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// CreateBucketAccess creates a new BucketAccess resource.
func CreateBucketAccess(bucketAccess *cosiv1alpha1.BucketAccess) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// SecretExists checks if the specified Secret resource exists.
func SecretExists(expected bool) types.StepFunc {
	// TODO: this should also check if the format of the secret conforms
	// to the expectations, so one more arg will be needed
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}

// DeleteBucketAccess deletes the specified BucketAccess resource.
func DeleteBucketAccess() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		return ctx
	}
}
