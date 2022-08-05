/* Copyright 2021 The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bucket

import (
	"context"
	"reflect"
	"testing"

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	fakebucketclientset "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned/fake"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
	fakespec "sigs.k8s.io/container-object-storage-interface-spec/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"

	"google.golang.org/grpc"
)

func TestInitializeKubeClient(t *testing.T) {
	client := fakekubeclientset.NewSimpleClientset()
	fakeDiscovery, ok := client.Discovery().(*fakediscovery.FakeDiscovery)
	if !ok {
		t.Fatalf("Couldn't convert Discovery() to *FakeDiscovery")
	}

	fakeVersion := &version.Info{
		GitVersion: "v1.0.0",
	}
	fakeDiscovery.FakedServerVersion = fakeVersion

	bl := BucketListener{}
	bl.InitializeKubeClient(client)

	if bl.kubeClient == nil {
		t.Errorf("KubeClient was nil")
	}

	expected := utilversion.MustParseSemantic(fakeVersion.GitVersion)
	if !reflect.DeepEqual(expected, bl.kubeVersion) {
		t.Errorf("Expected %+v, but got %+v", expected, bl.kubeVersion)
	}
}

func TestInitializeBucketClient(t *testing.T) {
	client := fakebucketclientset.NewSimpleClientset()

	bl := BucketListener{}
	bl.InitializeBucketClient(client)

	if bl.bucketClient == nil {
		t.Errorf("BucketClient was nil")
	}
}

func TestAddWrongProvisioner(t *testing.T) {
	provisioner := "provisioner1"
	mpc := struct{ fakespec.FakeProvisionerClient }{}
	mpc.FakeProvisionerCreateBucket = func(ctx context.Context,
		in *cosi.ProvisionerCreateBucketRequest,
		opts ...grpc.CallOption) (*cosi.ProvisionerCreateBucketResponse, error) {
		t.Errorf("grpc client called")
		return nil, nil
	}

	bl := BucketListener{
		provisionerName:   provisioner,
		provisionerClient: &mpc,
	}

	b := v1alpha1.Bucket{
		Spec: v1alpha1.BucketSpec{
			Provisioner: "provisioner2",
		},
	}
	ctx := context.TODO()
	err := bl.Add(ctx, &b)
	if err != nil {
		t.Errorf("Error returned: %+v", err)
	}
}

func TestDeleteWrongProvisioner(t *testing.T) {
	provisioner := "provisioner1"
	mpc := struct{ fakespec.FakeProvisionerClient }{}
	mpc.FakeProvisionerDeleteBucket = func(ctx context.Context,
		in *cosi.ProvisionerDeleteBucketRequest,
		opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error) {
		t.Errorf("grpc client called")
		return nil, nil
	}

	bl := BucketListener{
		provisionerName:   provisioner,
		provisionerClient: &mpc,
	}

	b := v1alpha1.Bucket{
		Spec: v1alpha1.BucketSpec{
			Provisioner: "provisioner2",
		},
	}
	ctx := context.TODO()
	err := bl.Delete(ctx, &b)
	if err != nil {
		t.Errorf("error returned: %+v", err)
	}
}

func TestBucketDeletion(t *testing.T) {
	provisioner := "provisioner1"
	bucketId := "bucket1"
	mpc := struct{ fakespec.FakeProvisionerClient }{}

	testCases := []struct {
		name       string
		setFields  func(*v1alpha1.Bucket)
		deleteFunc func(ctx context.Context,
			in *cosi.ProvisionerDeleteBucketRequest,
			opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error)
	}{
		{
			name: "BucketDeletion",
			setFields: func(b *v1alpha1.Bucket) {
				b.Status.BucketID = bucketId
			},
			deleteFunc: func(ctx context.Context,
				req *cosi.ProvisionerDeleteBucketRequest,
				opts ...grpc.CallOption) (*cosi.ProvisionerDeleteBucketResponse, error) {
				inBucketId := req.BucketId
				if inBucketId != bucketId {
					t.Errorf("expected %s, got %s", bucketId, inBucketId)
				}
				return &cosi.ProvisionerDeleteBucketResponse{}, nil
			},
		},
	}

	for _, tc := range testCases {
		b := v1alpha1.Bucket{
			Spec: v1alpha1.BucketSpec{
				Provisioner: provisioner,
			},
			Status: v1alpha1.BucketStatus{
				BucketAvailable: true,
			},
		}

		ctx := context.TODO()
		client := fakebucketclientset.NewSimpleClientset(&b)
		mpc.FakeProvisionerDeleteBucket = tc.deleteFunc
		bl := BucketListener{
			provisionerName:   provisioner,
			provisionerClient: &mpc,
			bucketClient:      client,
		}

		tc.setFields(&b)
		t.Logf(tc.name)
		err := bl.Delete(ctx, &b)
		if err != nil {
			t.Errorf("Error running TestBucketDeletion: %v", err)
		}

		updatedB, _ := client.ObjectstorageV1alpha1().Buckets().Get(ctx, b.Name, metav1.GetOptions{})
		if updatedB.Status.BucketAvailable != false {
			t.Errorf("Expected %t, got %t", false, b.Status.BucketAvailable)
		}
	}
}
