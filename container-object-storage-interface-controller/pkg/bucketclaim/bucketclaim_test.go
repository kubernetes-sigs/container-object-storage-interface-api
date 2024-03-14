package bucketclaim

import (
	"context"
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	types "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	fakebucketclientset "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned/fake"
	"sigs.k8s.io/container-object-storage-interface-api/controller/events"
	"sigs.k8s.io/container-object-storage-interface-controller/pkg/util"
)

var classGoldParameters = map[string]string{
	"param1": "value1",
	"param2": "value2",
}

var goldClass = types.BucketClass{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "objectstorage.k8s.io/v1alpha1",
		Kind:       "BucketClass",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "classgold",
	},
	DriverName:     "sample.cosi.driver",
	Parameters:     classGoldParameters,
	DeletionPolicy: types.DeletionPolicyDelete,
}

var bucketClaim1 = types.BucketClaim{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "objectstorage.k8s.io/v1alpha1",
		Kind:       "BucketClaim",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bucketclaim1",
		Namespace: "default",
		UID:       "12345-67890",
	},
	Spec: types.BucketClaimSpec{
		BucketClassName: "classgold",
		Protocols:       []types.Protocol{types.ProtocolAzure, types.ProtocolS3},
	},
}

var bucketClaim2 = types.BucketClaim{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "objectstorage.k8s.io/v1alpha1",
		Kind:       "BucketClaim",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bucketclaim2",
		Namespace: "default",
		UID:       "abcde-fghijk",
	},
	Spec: types.BucketClaimSpec{
		BucketClassName: "classgold",
		Protocols:       []types.Protocol{types.ProtocolAzure, types.ProtocolGCP},
	},
}

// Test basic add functionality
func TestAddBR(t *testing.T) {
	runCreateBucket(t)
}

// Test add with multipleBRs
func TestAddWithMultipleBR(t *testing.T) {
	runCreateBucketWithMultipleBR(t)
}

// Test add idempotency
func TestAddBRIdempotency(t *testing.T) {
	runCreateBucketIdempotency(t)
}

func runCreateBucket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := fakebucketclientset.NewSimpleClientset()
	kubeClient := fakekubeclientset.NewSimpleClientset()
	eventRecorder := record.NewFakeRecorder(3)

	listener := NewBucketClaimListener()
	listener.InitializeKubeClient(kubeClient)
	listener.InitializeBucketClient(client)
	listener.InitializeEventRecorder(eventRecorder)

	bucketclass, err := util.CreateBucketClass(ctx, client, &goldClass)
	if err != nil {
		t.Fatalf("Error occurred when creating BucketClass: %v", err)
	}

	bucketClaim, err := util.CreateBucketClaim(ctx, client, &bucketClaim1)
	if err != nil {
		t.Fatalf("Error occurred when creating BucketClaim: %v", err)
	}

	listener.Add(ctx, bucketClaim)

	bucketList := util.GetBuckets(ctx, client, 1)
	defer util.DeleteObjects(ctx, client, *bucketClaim, *bucketclass, bucketList.Items)

	if len(bucketList.Items) != 1 {
		t.Fatalf("Expecting a single Bucket created but found %v", len(bucketList.Items))
	}
	bucket := bucketList.Items[0]

	bucketClaim, err = client.ObjectstorageV1alpha1().BucketClaims(bucketClaim.Namespace).Get(ctx, bucketClaim.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error occurred when reading BucketClaim: %v", err)
	}

	if util.ValidateBucket(bucket, *bucketClaim, *bucketclass) {
		return
	} else {
		t.Fatalf("Failed to compare the resulting Bucket with the BucketClaim %v and BucketClass %v", bucketClaim, bucketclass)
	}
}

func runCreateBucketWithMultipleBR(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := fakebucketclientset.NewSimpleClientset()
	kubeClient := fakekubeclientset.NewSimpleClientset()
	eventRecorder := record.NewFakeRecorder(3)

	listener := NewBucketClaimListener()
	listener.InitializeKubeClient(kubeClient)
	listener.InitializeBucketClient(client)
	listener.InitializeEventRecorder(eventRecorder)

	bucketclass, err := util.CreateBucketClass(ctx, client, &goldClass)
	if err != nil {
		t.Fatalf("Error occurred when creating BucketClass: %v", err)
	}

	bucketClaim, err := util.CreateBucketClaim(ctx, client, &bucketClaim1)
	if err != nil {
		t.Fatalf("Error occurred when creating BucketClaim: %v", err)
	}

	bucketClaim2, err := util.CreateBucketClaim(ctx, client, &bucketClaim2)
	if err != nil {
		t.Fatalf("Error occurred when creating BucketClaim: %v", err)
	}

	listener.Add(ctx, bucketClaim)
	listener.Add(ctx, bucketClaim2)

	bucketList := util.GetBuckets(ctx, client, 2)
	defer util.DeleteObjects(ctx, client, *bucketClaim, *bucketClaim2, *bucketclass, bucketList.Items)
	if len(bucketList.Items) != 2 {
		t.Fatalf("Expecting two Buckets created but found %v", len(bucketList.Items))
	}
	bucket := bucketList.Items[0]
	bucket2 := bucketList.Items[1]

	bucketClaim, err = client.ObjectstorageV1alpha1().BucketClaims(bucketClaim.Namespace).Get(ctx, bucketClaim.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error occurred when reading BucketClaim: %v", err)
	}
	bucketClaim2, err = client.ObjectstorageV1alpha1().BucketClaims(bucketClaim2.Namespace).Get(ctx, bucketClaim2.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error occurred when reading BucketClaim: %v", err)
	}

	if (util.ValidateBucket(bucket, *bucketClaim, *bucketclass) && util.ValidateBucket(bucket2, *bucketClaim2, *bucketclass)) ||
		(util.ValidateBucket(bucket2, *bucketClaim, *bucketclass) && util.ValidateBucket(bucket, *bucketClaim2, *bucketclass)) {
		return
	} else {
		t.Fatalf("Failed to compare the resulting Bucket with the BucketClaim %v and BucketClass %v", bucketClaim, bucketclass)
	}
}

func runCreateBucketIdempotency(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := fakebucketclientset.NewSimpleClientset()
	kubeClient := fakekubeclientset.NewSimpleClientset()
	eventRecorder := record.NewFakeRecorder(3)

	listener := NewBucketClaimListener()
	listener.InitializeKubeClient(kubeClient)
	listener.InitializeBucketClient(client)
	listener.InitializeEventRecorder(eventRecorder)

	bucketclass, err := util.CreateBucketClass(ctx, client, &goldClass)
	if err != nil {
		t.Fatalf("Error occurred when creating BucketClass: %v", err)
	}

	bucketClaim, err := util.CreateBucketClaim(ctx, client, &bucketClaim1)
	if err != nil {
		t.Fatalf("Error occurred when creating BucketClaim: %v", err)
	}

	listener.Add(ctx, bucketClaim)

	bucketList := util.GetBuckets(ctx, client, 1)
	defer util.DeleteObjects(ctx, client, *bucketClaim, *bucketclass, bucketList.Items)

	if len(bucketList.Items) != 1 {
		t.Errorf("Expecting a single Bucket created but found %v", len(bucketList.Items))
	}
	bucket := bucketList.Items[0]

	bucketClaim, err = client.ObjectstorageV1alpha1().BucketClaims(bucketClaim.Namespace).Get(ctx, bucketClaim.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error occurred when reading BucketClaim: %v", err)
	}

	if util.ValidateBucket(bucket, *bucketClaim, *bucketclass) {
		return
	} else {
		t.Fatalf("Failed to compare the resulting Bucket with the BucketClaim %v and BucketClass %v", bucketClaim, bucketclass)
		// call the add directly the second time
	}

	listener.Add(ctx, bucketClaim)

	bucketList = util.GetBuckets(ctx, client, 1)
	if len(bucketList.Items) != 1 {
		t.Fatalf("Expecting a single Bucket created but found %v", len(bucketList.Items))
	}
}

// Test recording events
func TestRecordEvents(t *testing.T) {
	t.Parallel()

	defaultBucketClaim := &v1alpha1.BucketClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bucketClaim",
			Namespace: "test-ns",
		},
		Spec: v1alpha1.BucketClaimSpec{
			BucketClassName: "test-bucketClass",
		},
	}

	for _, tc := range []struct {
		name          string
		expectedEvent string
		eventTrigger  func(*testing.T, *BucketClaimListener)
	}{
		{
			name: "ExistingBucketNotFound",
			expectedEvent: newEvent(
				v1.EventTypeWarning,
				events.FailedCreateBucket,
				"buckets.objectstorage.k8s.io \"existing-bucket\" not found"),
			eventTrigger: func(t *testing.T, bcl *BucketClaimListener) {
				ctx := context.TODO()

				bucketClaim := defaultBucketClaim.DeepCopy()
				bucketClaim.Spec.ExistingBucketName = "existing-bucket"

				err := bcl.Add(ctx, bucketClaim)
				if !kubeerrors.IsNotFound(err) {
					t.Errorf("expected Not Found error got %v", err)
				}
			},
		},
		{
			name: "BucketClassNotFound",
			expectedEvent: newEvent(
				v1.EventTypeWarning,
				events.FailedCreateBucket,
				"bucketclasses.objectstorage.k8s.io \"test-bucketClass\" not found"),
			eventTrigger: func(t *testing.T, listener *BucketClaimListener) {
				ctx := context.TODO()
				bucketClaim := defaultBucketClaim.DeepCopy()

				err := listener.Add(ctx, bucketClaim)
				if !kubeerrors.IsNotFound(err) {
					t.Errorf("expected Not Found error got %v", err)
				}
			},
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fakebucketclientset.NewSimpleClientset()
			kubeClient := fakekubeclientset.NewSimpleClientset()
			eventRecorder := record.NewFakeRecorder(1)

			listener := NewBucketClaimListener()
			listener.InitializeKubeClient(kubeClient)
			listener.InitializeBucketClient(client)
			listener.InitializeEventRecorder(eventRecorder)

			tc.eventTrigger(t, listener)

			select {
			case event, ok := <-eventRecorder.Events:
				if ok {
					if event != tc.expectedEvent {
						t.Errorf("expected %s got %s", tc.expectedEvent, event)
					}
				} else {
					t.Error("channel closed, no event")
				}
			default:
				t.Errorf("no event after trigger")
			}
		})
	}
}

func newEvent(eventType, reason, message string) string {
	return fmt.Sprintf("%s %s %s", eventType, reason, message)
}
