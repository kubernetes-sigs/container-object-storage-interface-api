package envfuncs

import (
	"context"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

// InstallCRDs installs the necessary CRDs unless skipping is specified.
func InstallCRDs(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

// UninstallCRDs uninstalls the necessary CRDs unless skipping is specified.
func UninstallCRDs(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
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
