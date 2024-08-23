package envfuncs

import (
	"context"
	"e2e/retry"
	"fmt"
	"testing"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
	"sigs.k8s.io/e2e-framework/support/kind"
)

type (
	ClusterCtxKey string
)

const clusterKey = ClusterCtxKey("cluster")

// CreateCluster creates a new Kubernetes cluster unless skipping is specified.
func CreateCluster(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		name := envconf.RandomName("cosi-e2e-cluster", 32)
		cluster := kind.NewCluster(name)

		kubeconfig, err := cluster.Create(ctx)
		if err != nil {
			return ctx, err
		}

		// stall a bit to allow most pods to come up
		ctx, err = IsClusterReady()(ctx, cfg)
		if err != nil {
			return ctx, err
		}

		// update environment with kubecofig file
		cfg.WithKubeconfigFile(kubeconfig)

		// propagate cluster value
		return context.WithValue(ctx, clusterKey, cluster), nil
	}
}

// IsClusterReady checks if the Kubernetes cluster is ready.
func IsClusterReady() types.EnvFunc {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		clientset, err := kubernetes.NewForConfig(cfg.Client().RESTConfig())
		if err != nil {
			return ctx, fmt.Errorf("failed to create clientset from klient: %w", err)
		}

		r := retry.NewLinearBackoffRetry()

		return ctx, r.Retry(func() error {
			_, err := clientset.ServerVersion()
			return err
		})
	}
}

func IsClusterReadyTest() types.TestEnvFunc {
	return func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		return IsClusterReady()(ctx, cfg)
	}
}

// DeleteCluster deletes the Kubernetes cluster unless skipping is specified.
func DeleteCluster(skip bool) types.EnvFunc {
	if skip {
		return Noop()
	}

	return func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
		cluster := ctx.Value(clusterKey).(*kind.Cluster)
		if cluster == nil {
			return ctx, fmt.Errorf("error getting kind cluster from context")
		}

		if err := cluster.Destroy(ctx); err != nil {
			return ctx, err
		}

		return ctx, nil
	}
}
