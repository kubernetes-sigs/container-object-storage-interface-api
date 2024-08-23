package assesments

import (
	"context"
	"e2e/retry"
	"fmt"
	"testing"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	cosi "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"

	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

// CRDsInstalled checks if the required COSI CRDs are installed in the cluster.
func CRDsInstalled() types.TestEnvFunc {
	return func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		var crds apiextensions.CustomResourceDefinitionList

		expectedCRDs := []string{
			"bucketaccessclasses.objectstorage.k8s.io",
			"bucketaccesses.objectstorage.k8s.io",
			"bucketclaims.objectstorage.k8s.io",
			"bucketclasses.objectstorage.k8s.io",
			"buckets.objectstorage.k8s.io",
		}

		if err := cfg.Client().Resources().List(ctx, &crds); err != nil {
			t.Error(err)
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
			t.Error("COSI CRDs not installed")
		}

		return ctx, nil
	}
}

// ObjectstorageControllerInstalled checks if the COSI object storage controller deployment is installed.
func ObjectstorageControllerInstalled() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var deploymentList apps.DeploymentList

		selector := resources.WithLabelSelector("app.kubernetes.io/part-of=container-object-storage-interface")

		if err := cfg.Client().Resources().List(ctx, &deploymentList, selector); err != nil {
			t.Error(err)
		}

		if len(deploymentList.Items) == 0 {
			t.Error("deployment not found")
		}

		return ctx
	}
}

// BucketAccessStatusGranted checks if the status of BucketAccess is granted.
func BucketAccessStatusGranted(granted bool) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketAccess, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketAccess)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		r := retry.NewLinearBackoffRetry()

		if err := r.Retry(func() error {
			actualBucketAccess := &cosi.BucketAccess{}

			err := cfg.Client().Resources().Get(ctx,
				bucketAccess.Name,
				bucketAccess.Namespace,
				actualBucketAccess,
			)

			if err != nil {
				return err
			}

			if actualBucketAccess.Status.AccessGranted != granted {
				return fmt.Errorf("expected: %v, actual: %v",
					granted,
					actualBucketAccess.Status.AccessGranted,
				)
			}

			return nil
		}); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// BucketClaimStatusReady checks if the status of BucketClaim is ready.
func BucketClaimStatusReady(ready bool) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketClaim, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketClaim)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		r := retry.NewLinearBackoffRetry()

		if err := r.Retry(func() error {
			actualBucketClaim := &cosi.BucketClaim{}

			err := cfg.Client().Resources().Get(ctx,
				bucketClaim.Name,
				bucketClaim.Namespace,
				actualBucketClaim,
			)

			if err != nil {
				return err
			}

			if actualBucketClaim.Status.BucketReady != ready {
				return fmt.Errorf("expected: %v, actual: %v",
					ready,
					actualBucketClaim.Status.BucketReady,
				)
			}

			return err
		}); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// CreateBucket creates a new Bucket resource.
func CreateBucket() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucket, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.Bucket)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		if err := cfg.Client().Resources().Create(ctx, bucket); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// BucketExists checks if the Bucket resource exists.
func BucketExists(expected bool) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketClaim, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketClaim)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		r := retry.NewLinearBackoffRetry()

		if err := r.Retry(func() error {
			bucket := &cosi.Bucket{}

			err := cfg.Client().Resources().Get(ctx,
				bucketClaim.Status.BucketName,
				bucketClaim.Namespace,
				bucket,
			)

			if errors.IsNotFound(err) {
				if expected {
					return err
				}
				// else ignore error
			} else if err != nil {
				return err
			}

			return nil
		}); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// CreateBucketClaim creates a new BucketClaim resource.
func CreateBucketClaim() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketClaim, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketClaim)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		if err := cfg.Client().Resources().Create(ctx, bucketClaim); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// DeleteBucketClaim deletes the specified BucketClaim resource.
func DeleteBucketClaim() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketClaim, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketClaim)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		if err := cfg.Client().Resources().Delete(ctx, bucketClaim); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// CreateBucketAccess creates a new BucketAccess resource.
func CreateBucketAccess() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketAccess, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketAccess)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		if err := cfg.Client().Resources().Create(ctx, bucketAccess); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// SecretExists checks if the specified Secret resource exists.
func SecretExists(expected bool) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketAccess, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketAccess)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		r := retry.NewLinearBackoffRetry()

		if err := r.Retry(func() error {
			// TODO: this should also check if the format of the secret conforms
			// to the expectations, so one more arg will be needed
			secret := &core.Secret{}

			err := cfg.Client().Resources().Get(ctx,
				bucketAccess.Spec.CredentialsSecretName,
				bucketAccess.Namespace,
				secret,
			)

			if errors.IsNotFound(err) {
				if expected {
					return err
				}
				// else ignore error
			} else if err != nil {
				return err
			}

			return nil
		}); err != nil {
			t.Error(err)
		}

		return ctx
	}
}

// DeleteBucketAccess deletes the specified BucketAccess resource.
func DeleteBucketAccess() types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		bucketAccess, ok := ctx.Value(BucketAccessTypeCtxKey).(*cosi.BucketAccess)
		if !ok {
			t.Error(ErrKeyNotFound)
			return ctx
		}

		if err := cfg.Client().Resources().Delete(ctx, bucketAccess); err != nil {
			t.Error(err)
		}

		return ctx
	}
}
