package envfuncs

import (
	"context"
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

type (
	TypeCtxKey string
)

const (
	CustomResourceDefinitionListTypeCtxKey = TypeCtxKey("apiextensions.CustomResourceDefinitionList")
)

func RegisterResources(objects ...runtime.Object) types.EnvFunc {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		for _, obj := range objects {
			switch typedObj := obj.(type) {
			case *apiextensions.CustomResourceDefinitionList:
				ctx = context.WithValue(ctx, CustomResourceDefinitionListTypeCtxKey, typedObj)

			default:
				panic(fmt.Sprintf("unsupported type: %T (Kind: %s)", typedObj, obj.GetObjectKind()))
			}
		}

		return ctx, nil
	}
}

// InstallCRDs installs the necessary CRDs unless skipping is specified.
func InstallCRDs(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		crds := ctx.Value(CustomResourceDefinitionListTypeCtxKey).(*apiextensions.CustomResourceDefinitionList)

		for _, crd := range crds.Items {
			if err := cfg.Client().Resources().Create(ctx, &crd); err != nil {
				return ctx, err
			}
		}

		return ctx, nil
	}
}

// UninstallCRDs uninstalls the necessary CRDs unless skipping is specified.
func UninstallCRDs(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		crds := ctx.Value(CustomResourceDefinitionListTypeCtxKey).(*apiextensions.CustomResourceDefinitionList)

		for _, crd := range crds.Items {
			if err := cfg.Client().Resources().Delete(ctx, &crd); err != nil {
				return ctx, err
			}
		}

		return ctx, nil
	}
}

// InstallController installs the controller unless skipping is specified.
func InstallController(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// UninstallController uninstalls the controller unless skipping is specified.
func UninstallController(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// InstallDriver installs the driver unless skipping is specified.
func InstallDriver(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// UninstallDriver uninstalls the driver unless skipping is specified.
func UninstallDriver(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// CreateBucketClass creates a BucketClass unless skipping is specified.
func CreateBucketClass(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// DeleteBucketClass deletes a BucketClass unless skipping is specified.
func DeleteBucketClass(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// CreateBucketAccessClass creates a BucketAccessClass unless skipping is specified.
func CreateBucketAccessClass(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// DeleteBucketAccessClass deletes a BucketAccessClass unless skipping is specified.
func DeleteBucketAccessClass(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}
