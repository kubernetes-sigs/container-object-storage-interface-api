package setup

import (
	"context"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

type (
	TypeCtxKey string
)

const (
	BucketTypeCtxKey       = TypeCtxKey("cosi.Bucket")
	BucketClaimTypeCtxKey  = TypeCtxKey("cosi.BucketClaim")
	BucketAccessTypeCtxKey = TypeCtxKey("cosi.BucketAccess")
)

func RegisterResourcesForTest(objects ...runtime.Object) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		for _, obj := range objects {
			switch typedObj := obj.(type) {
			case *v1alpha1.Bucket:
				ctx = context.WithValue(ctx, BucketTypeCtxKey, typedObj)

			case *v1alpha1.BucketClaim:
				ctx = context.WithValue(ctx, BucketClaimTypeCtxKey, typedObj)

			case *v1alpha1.BucketAccess:
				ctx = context.WithValue(ctx, BucketAccessTypeCtxKey, typedObj)

			default:
				panic(fmt.Sprintf("invalid type: %T (Kind: %s)", typedObj, obj.GetObjectKind()))
			}
		}

		return ctx
	}
}
