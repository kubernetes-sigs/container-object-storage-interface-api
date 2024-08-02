package cositest

import (
	"e2e/assesments/cosi"
	"e2e/envfuncs"
	"flag"
	"log"
	"os"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	cosiv1alpha1 "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
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

	foo string
)

func init() {
	apiextensionsv1.AddToScheme(scheme.Scheme)
	cosiv1alpha1.AddToScheme(scheme.Scheme)
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

	testenv.Setup(
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
		cosi.CRDsInstalled(),
	)

	testenv.AfterEachTest(
		envfuncs.DeleteNSForTest(),
	)

	os.Exit(testenv.Run(m))
}

func TestBucketProvisioning(t *testing.T) {
	testenv.Test(t,
		features.New("Greenfield Bucket").
			Assess("BucketClaim is created",
				cosi.CreateBucketClaim(&v1alpha1.BucketClaim{})).
			Assess("Bucket is created",
				cosi.CreateBucket(&v1alpha1.Bucket{})).
			Assess("BucketClaim has ready status",
				cosi.BucketClaimStatusReady(true)).
			Assess("BucketClaim is deleted",
				cosi.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				cosi.BucketExists(false)).
			Feature(),

		features.New("Brownfield Bucket").
			Assess("BucketClaim is created",
				cosi.CreateBucketClaim(&v1alpha1.BucketClaim{})).
			Assess("Bucket is created",
				cosi.CreateBucket(&v1alpha1.Bucket{})).
			Assess("BucketClaim has ready status",
				cosi.BucketClaimStatusReady(true)).
			Assess("BucketClaim is deleted",
				cosi.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				cosi.BucketExists(false)).
			Feature(),
	)
}

func TestBucketAccessProvisioning(t *testing.T) {
	testenv.Test(t,
		features.New("BucketAccess for S3").
			Assess("BucketClaim is created",
				cosi.CreateBucketClaim(&v1alpha1.BucketClaim{})).
			Assess("Bucket is created",
				cosi.CreateBucket(&v1alpha1.Bucket{})).
			Assess("BucketClaim has ready status",
				cosi.BucketClaimStatusReady(true)).
			Assess("BucketAccess is created",
				cosi.CreateBucketAccess(&v1alpha1.BucketAccess{})).
			Assess("BucketAccess has ready status",
				cosi.BucketAccessStatusReady(true)).
			Assess("Secret is created",
				cosi.SecretExists(true)).
			Assess("BucketAccess is deleted",
				cosi.DeleteBucketAccess()).
			Assess("Secret is deleted",
				cosi.SecretExists(false)).
			Assess("BucketClaim is deleted",
				cosi.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				cosi.BucketExists(false)).
			Feature(),

		features.New("BucketAccess for Azure").
			Assess("BucketClaim is created",
				cosi.CreateBucketClaim(&v1alpha1.BucketClaim{})).
			Assess("Bucket is created",
				cosi.CreateBucket(&v1alpha1.Bucket{})).
			Assess("BucketClaim has ready status",
				cosi.BucketClaimStatusReady(true)).
			Assess("BucketAccess is created",
				cosi.CreateBucketAccess(&v1alpha1.BucketAccess{})).
			Assess("BucketAccess has ready status",
				cosi.BucketAccessStatusReady(true)).
			Assess("Secret is created",
				cosi.SecretExists(true)).
			Assess("BucketAccess is deleted",
				cosi.DeleteBucketAccess()).
			Assess("Secret is deleted",
				cosi.SecretExists(false)).
			Assess("BucketClaim is deleted",
				cosi.DeleteBucketClaim()).
			Assess("Bucket is deleted",
				cosi.BucketExists(false)).
			Feature(),
	)
}
