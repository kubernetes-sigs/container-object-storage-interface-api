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

package bucketaccess

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1"
	fakebucketclientset "sigs.k8s.io/container-object-storage-interface-api/clientset/fake"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
	fakespec "sigs.k8s.io/container-object-storage-interface-spec/fake"
)

func TestInitializeKubeClient(t *testing.T) {
	client := fakekubeclientset.NewSimpleClientset()
	fakeDiscovery, ok := client.Discovery().(*fakediscovery.FakeDiscovery)
	if !ok {
		t.Fatalf("couldn't convert Discovery() to *FakeDiscovery")
	}

	fakeVersion := &version.Info{
		GitVersion: "v1.0.0",
	}
	fakeDiscovery.FakedServerVersion = fakeVersion

	bal := BucketAccessListener{}
	bal.InitializeKubeClient(client)

	if bal.kubeClient == nil {
		t.Errorf("KubeClient was nil")
	}

	expected := utilversion.MustParseSemantic(fakeVersion.GitVersion)
	if !reflect.DeepEqual(expected, bal.kubeVersion) {
		t.Errorf("Expected %+v, but got %+v", expected, bal.kubeVersion)
	}
}

func TestInitializeBucketClient(t *testing.T) {
	client := fakebucketclientset.NewSimpleClientset()

	bal := BucketAccessListener{}
	bal.InitializeBucketClient(client)

	if bal.bucketClient == nil {
		t.Errorf("BucketClient not initialized, expected not nil")
	}
}

func TestAddWrongProvisioner(t *testing.T) {
	provisioner := "provisioner1"
	bucketName := "bucket1"
	bucketId := "bucketId1"
	accountId := "accountId1"
	bucketAccessRequestName := "bar1"
	policy := "policy1"

	mpc := struct{ fakespec.FakeProvisionerClient }{}
	mpc.FakeProvisionerGrantBucketAccess = func(ctx context.Context,
		in *cosi.ProvisionerGrantBucketAccessRequest,
		opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error) {
		t.Errorf("grpc client called")
		return &cosi.ProvisionerGrantBucketAccessResponse{
			AccountId: accountId,
		}, nil
	}

	b := v1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name: bucketName,
		},
		Spec: v1alpha1.BucketSpec{
			Provisioner: provisioner + "-invalid",
			Protocol:    v1alpha1.Protocol{},
		},
		Status: v1alpha1.BucketStatus{
			BucketID: bucketId,
		},
	}

	ba := v1alpha1.BucketAccess{
		Spec: v1alpha1.BucketAccessSpec{
			BucketName: bucketName,
			BucketAccessRequest: &corev1.ObjectReference{
				Name: bucketAccessRequestName,
			},
			PolicyActionsConfigMapData: policy,
		},
	}
	client := fakebucketclientset.NewSimpleClientset(&ba, &b)
	kubeClient := fakekubeclientset.NewSimpleClientset()
	bal := BucketAccessListener{
		provisionerName:   provisioner,
		provisionerClient: &mpc,
		bucketClient:      client,
		kubeClient:        kubeClient,
	}

	ctx := context.TODO()
	err := bal.Add(ctx, &ba)
	if err != nil {
		t.Errorf("Error returned: %+v", err)
	}
}

func TestAddBucketAccess(t *testing.T) {
	provisioner := "provisioner"
	bucketName := "bucket1"
	bucketId := "bucketId1"
	bucketAccessRequestName := "bar1"

	policy := "policy1"
	accountId := "account1"
	creds := "credsContents"
	ns := "testns"
	mpc := struct{ fakespec.FakeProvisionerClient }{}

	testCases := []struct {
		name      string
		setFields func(ba *v1alpha1.BucketAccess)
		grantFunc func(ctx context.Context,
			in *cosi.ProvisionerGrantBucketAccessRequest,
			opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error)
	}{
		{
			name: "TestAddBucketAccess",
			setFields: func(ba *v1alpha1.BucketAccess) {

			},
			grantFunc: func(ctx context.Context,
				req *cosi.ProvisionerGrantBucketAccessRequest,
				opts ...grpc.CallOption) (*cosi.ProvisionerGrantBucketAccessResponse, error) {

				return &cosi.ProvisionerGrantBucketAccessResponse{
					AccountId:   accountId,
					Credentials: creds,
				}, nil
			},
		},
	}

	for _, tc := range testCases {
		b := v1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Name: bucketName,
			},
			Spec: v1alpha1.BucketSpec{
				Provisioner: provisioner,
				Protocol:    v1alpha1.Protocol{},
			},
			Status: v1alpha1.BucketStatus{
				BucketID: bucketId,
			},
		}

		ba := v1alpha1.BucketAccess{
			Spec: v1alpha1.BucketAccessSpec{
				BucketName: bucketName,
				BucketAccessRequest: &corev1.ObjectReference{
					Name: bucketAccessRequestName,
				},
				PolicyActionsConfigMapData: policy,
			},
		}

		ctx := context.TODO()
		tc.setFields(&ba)

		client := fakebucketclientset.NewSimpleClientset(&ba, &b)
		kubeClient := fakekubeclientset.NewSimpleClientset()
		mpc.FakeProvisionerGrantBucketAccess = tc.grantFunc

		bal := BucketAccessListener{
			provisionerName:   provisioner,
			provisionerClient: &mpc,
			bucketClient:      client,
			kubeClient:        kubeClient,
			namespace:         ns,
		}

		t.Logf(tc.name)
		err := bal.Add(ctx, &ba)
		if err != nil {
			t.Errorf("Add returned: %+v", err)
		}

		updatedBA, _ := bal.BucketAccesses().Get(ctx, ba.Name, metav1.GetOptions{})
		if updatedBA.Status.AccessGranted != true {
			t.Errorf("Expected %t, got %t", true, ba.Status.AccessGranted)
		}
		if !strings.EqualFold(updatedBA.Status.AccountID, accountId) {
			t.Errorf("Expected %s, got %s", accountId, updatedBA.Status.AccountID)
		}

		secretName := "ba-" + string(ba.UID)
		secret, err := bal.Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("minted secret creation failed: %v", err)
		}

		if secret.StringData["Credentials"] != creds {
			t.Errorf("Expected %s, got %s",
				creds,
				secret.StringData["Credentials"])
		}
	}
}
