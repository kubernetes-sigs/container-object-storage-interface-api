// Package envfuncs provides a collection of functions that return closures
// conforming to the types defined in sigs.k8s.io/e2e-framework/pkg/types .
// These functions are designed to be used within the Kubernetes end-to-end
// testing framework to facilitate the creation and management of test
// environments.

package envfuncs

import (
	"context"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// Noop returns an EnvFunc that performs no operation. This can be used as a placeholder
// or default function within an environment configuration.
func Noop() func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}
