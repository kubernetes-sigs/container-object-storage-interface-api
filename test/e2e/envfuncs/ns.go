package envfuncs

import (
	"context"
	"fmt"
	"testing"

	"e2e/envfuncs/helpers"

	core "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

// CreateNSForTest creates a random namespace with the runID as a prefix. It is stored in the context
// so that the deleteNSForTest routine can look it up and delete it.
func CreateNSForTest() types.TestEnvFunc {
	return func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		ns := envconf.RandomName("e2e", 16)
		ctx = context.WithValue(ctx, helpers.GetNamespaceKey(t), ns)

		t.Logf("Creating NS %v for test %v", ns, t.Name())
		nsObj := core.Namespace{}
		nsObj.Name = ns
		return ctx, cfg.Client().Resources().Create(ctx, &nsObj)
	}
}

// DeleteNSForTest looks up the namespace corresponding to the given test and deletes it.
func DeleteNSForTest() types.TestEnvFunc {
	return func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		ns := fmt.Sprint(ctx.Value(helpers.GetNamespaceKey(t)))
		t.Logf("Deleting NS %v for test %v", ns, t.Name())

		nsObj := core.Namespace{}
		nsObj.Name = ns
		return ctx, cfg.Client().Resources().Delete(ctx, &nsObj)
	}
}
