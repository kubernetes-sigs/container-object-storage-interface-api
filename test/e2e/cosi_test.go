package cositest

import (
	"e2e/assesments"
	"e2e/envfuncs"
	"e2e/envfuncs/helpers"
	"flag"
	"log"
	"os"
	"testing"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	cosi "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	testenv env.Environment

	noKinD                bool
	noInstallCRDs         bool
	noInstallController   bool
	noInstallSampleDriver bool
)

func init() {
	apiextensions.AddToScheme(scheme.Scheme)
	cosi.AddToScheme(scheme.Scheme)
}

func TestMain(m *testing.M) {
	flag.BoolVar(
		&noKinD,
		"no-kind",
		false,
		"Start new environment with kind",
	)
	flag.BoolVar(
		&noInstallCRDs,
		"cosi.no-install-crds",
		false,
		"Disable installing CRDs on cluster",
	)
	flag.BoolVar(
		&noInstallController,
		"cosi.no-install-controller",
		false,
		"Disable installing COSI Controller on cluster",
	)
	flag.BoolVar(
		&noInstallSampleDriver,
		"cosi.no-install-sample-driver",
		false,
		"Disable installing sample driver on cluster",
	)
	flag.Parse()

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("failed to build envconf from flags: %s", err)
	}
	testenv = env.NewWithConfig(cfg)

	crds := &apiextensions.CustomResourceDefinitionList{}
	if !noInstallCRDs {
		crdPaths := []string{
			"../../client/config/crd/objectstorage.k8s.io_bucketaccessclasses.yaml",
			"../../client/config/crd/objectstorage.k8s.io_bucketaccesses.yaml",
			"../../client/config/crd/objectstorage.k8s.io_bucketclaims.yaml",
			"../../client/config/crd/objectstorage.k8s.io_bucketclasses.yaml",
			"../../client/config/crd/objectstorage.k8s.io_buckets.yaml",
		}

		for _, path := range crdPaths {

			crd, err := helpers.Load(path)
			if err != nil {
				log.Fatalf("failed to load resource: %s", err)
			}

			crds.Items = append(crds.Items, *crd)
		}

	}

	testenv.Setup(
		envfuncs.RegisterResources(crds),
		envfuncs.CreateCluster(noKinD),

		envfuncs.InstallCRDs(noInstallCRDs),
		envfuncs.InstallController(noInstallController),
		envfuncs.InstallDriver(noInstallSampleDriver),

		envfuncs.CreateBucketClass(false),
		envfuncs.CreateBucketAccessClass(false),
	)

	testenv.Finish(
		envfuncs.DeleteBucketAccessClass(false),
		envfuncs.DeleteBucketClass(false),

		envfuncs.UninstallDriver(noInstallSampleDriver),
		envfuncs.UninstallController(noInstallController),
		envfuncs.UninstallCRDs(noInstallCRDs),

		envfuncs.DeleteCluster(noKinD),
	)

	testenv.BeforeEachTest(
		envfuncs.IsClusterReadyTest(),
		envfuncs.CreateNSForTest(),
		assesments.CRDsInstalled(),
	)

	testenv.AfterEachTest(
		envfuncs.DeleteNSForTest(),
	)

	os.Exit(testenv.Run(m))
}

func TestBucketProvisioning(t *testing.T) {
	testenv.Test(t,
		features.New("Greenfield Bucket").
			Assess("Resources are registered",
				assesments.RegisterResourcesForTest(
					&cosi.BucketClaim{},
				)).
			Assess("BucketClaim is created",
				assesments.CreateBucketClaim()).
			Assess("Bucket is created",
				assesments.BucketExists(true)).
			Assess("BucketClaim has ready status",
				assesments.BucketClaimStatusReady(true)).
			Assess("BucketClaim is deleted",
				assesments.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				assesments.BucketExists(false)).
			Feature(),

		features.New("Brownfield Bucket").
			Assess("Resources are registered",
				assesments.RegisterResourcesForTest(
					&cosi.Bucket{},
					&cosi.BucketClaim{},
				)).
			Assess("BucketClaim is created",
				assesments.CreateBucketClaim()).
			Assess("Bucket is created",
				assesments.CreateBucket()).
			Assess("BucketClaim has ready status",
				assesments.BucketClaimStatusReady(true)).
			Assess("BucketClaim is deleted",
				assesments.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				assesments.BucketExists(false)).
			Feature(),
	)
}

func TestBucketAccessProvisioning(t *testing.T) {
	testenv.Test(t,
		features.New("BucketAccess for S3").
			Assess("Resources are registered",
				assesments.RegisterResourcesForTest(
					&cosi.BucketClaim{},
					&cosi.BucketAccess{},
				)).
			Assess("BucketClaim is created",
				assesments.CreateBucketClaim()).
			Assess("Bucket is created",
				assesments.BucketExists(true)).
			Assess("BucketClaim has ready status",
				assesments.BucketClaimStatusReady(true)).
			Assess("BucketAccess is created",
				assesments.CreateBucketAccess()).
			Assess("BucketAccess has ready status",
				assesments.BucketAccessStatusGranted(true)).
			Assess("Secret is created",
				assesments.SecretExists(true)).
			Assess("BucketAccess is deleted",
				assesments.DeleteBucketAccess()).
			Assess("Secret is deleted",
				assesments.SecretExists(false)).
			Assess("BucketClaim is deleted",
				assesments.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				assesments.BucketExists(false)).
			Feature(),

		features.New("BucketAccess for Azure").
			Assess("Resources are registered",
				assesments.RegisterResourcesForTest(
					&cosi.BucketClaim{},
					&cosi.BucketAccess{},
				)).
			Assess("BucketClaim is created",
				assesments.CreateBucketClaim()).
			Assess("Bucket is created",
				assesments.BucketExists(true)).
			Assess("BucketClaim has ready status",
				assesments.BucketClaimStatusReady(true)).
			Assess("BucketAccess is created",
				assesments.CreateBucketAccess()).
			Assess("BucketAccess has ready status",
				assesments.BucketAccessStatusGranted(true)).
			Assess("Secret is created",
				assesments.SecretExists(true)).
			Assess("BucketAccess is deleted",
				assesments.DeleteBucketAccess()).
			Assess("Secret is deleted",
				assesments.SecretExists(false)).
			Assess("BucketClaim is deleted",
				assesments.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				assesments.BucketExists(false)).
			Feature(),
	)
}
