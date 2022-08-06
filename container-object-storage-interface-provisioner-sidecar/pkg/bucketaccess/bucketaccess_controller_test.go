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
	"time"

	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	fakebucketclientset "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned/fake"
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
	driver := "driver1"
	bucketAccessClassName := "bucketAccessClass1"
	bucketClaimName := "bucketClaim1"
	accountId := "accountId1"
	secretName := "secret1"
	timeNow := time.Now()
	secret := map[string]string{
		"accessToken": "randomValue",
		"expiryTs":    timeNow.String(),
	}

	credentialDetails := &cosi.CredentialDetails{
		Secrets: secret,
	}
	credential := map[string]*cosi.CredentialDetails{
		"azure": credentialDetails,
	}

	mpc := struct{ fakespec.FakeProvisionerClient }{}
	mpc.FakeDriverGrantBucketAccess = func(ctx context.Context,
		in *cosi.DriverGrantBucketAccessRequest,
		opts ...grpc.CallOption) (*cosi.DriverGrantBucketAccessResponse, error) {
		t.Errorf("grpc client called")
		return &cosi.DriverGrantBucketAccessResponse{
			AccountId:   accountId,
			Credentials: credential,
		}, nil
	}

	b := v1alpha1.BucketAccessClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: bucketAccessClassName,
		},
		DriverName:         driver + "-invalid",
		AuthenticationType: v1alpha1.AuthenticationTypeKey,
	}

	ba := v1alpha1.BucketAccess{
		Spec: v1alpha1.BucketAccessSpec{
			BucketClaimName:       bucketClaimName,
			Protocol:              v1alpha1.ProtocolAzure,
			BucketAccessClassName: bucketAccessClassName,
			CredentialsSecretName: secretName,
		},
	}

	client := fakebucketclientset.NewSimpleClientset(&ba, &b)
	kubeClient := fakekubeclientset.NewSimpleClientset()
	bal := BucketAccessListener{
		driverName:        driver,
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
	driver := "driver"
	bucketName := "bucket1"
	bucketId := "bucketId1"
	bucketClaimName := "bucketClaim1"
	bucketClassName := "bucketClass1"
	bucketAccessClassName := "bucketAccessClass1"
	bucketAccessName := "bucketAccess1"
	secretName := "secret1"

	accountId := "account1"
	ns := "testns"

	timeNow := time.Now()
	secret := map[string]string{
		"accessToken": "randomValue",
		"expiryTs":    timeNow.String(),
	}

	credentialDetails := &cosi.CredentialDetails{
		Secrets: secret,
	}
	creds := map[string]*cosi.CredentialDetails{
		"azure": credentialDetails,
	}

	mpc := struct{ fakespec.FakeProvisionerClient }{}

	testCases := []struct {
		name      string
		setFields func(ba *v1alpha1.BucketAccess)
		grantFunc func(ctx context.Context,
			in *cosi.DriverGrantBucketAccessRequest,
			opts ...grpc.CallOption) (*cosi.DriverGrantBucketAccessResponse, error)
	}{
		{
			name: "TestAddBucketAccess",
			setFields: func(ba *v1alpha1.BucketAccess) {

			},
			grantFunc: func(ctx context.Context,
				req *cosi.DriverGrantBucketAccessRequest,
				opts ...grpc.CallOption) (*cosi.DriverGrantBucketAccessResponse, error) {

				return &cosi.DriverGrantBucketAccessResponse{
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
				DriverName:     driver,
				Protocols:      []v1alpha1.Protocol{v1alpha1.ProtocolAzure},
				DeletionPolicy: v1alpha1.DeletionPolicyRetain,
			},
			Status: v1alpha1.BucketStatus{
				BucketID:    bucketId,
				BucketReady: true,
			},
		}

		bc := v1alpha1.BucketClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bucketClaimName,
				Namespace: ns,
			},
			Spec: v1alpha1.BucketClaimSpec{
				BucketClassName: bucketClassName,
				Protocols:       []v1alpha1.Protocol{v1alpha1.ProtocolAzure},
			},
			Status: v1alpha1.BucketClaimStatus{
				BucketReady: true,
				BucketName:  bucketName,
			},
		}

		bac := v1alpha1.BucketAccessClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: bucketAccessClassName,
			},
			DriverName:         driver,
			AuthenticationType: v1alpha1.AuthenticationTypeKey,
		}

		ba := v1alpha1.BucketAccess{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bucketAccessName,
				Namespace: ns,
			},
			Spec: v1alpha1.BucketAccessSpec{
				BucketClaimName:       bucketClaimName,
				Protocol:              v1alpha1.ProtocolAzure,
				BucketAccessClassName: bucketAccessClassName,
				CredentialsSecretName: secretName,
			},
		}

		ctx := context.TODO()
		tc.setFields(&ba)

		client := fakebucketclientset.NewSimpleClientset(&b, &bc, &bac, &ba)
		kubeClient := fakekubeclientset.NewSimpleClientset()
		mpc.FakeDriverGrantBucketAccess = tc.grantFunc

		bal := BucketAccessListener{
			driverName:        driver,
			provisionerClient: &mpc,
			bucketClient:      client,
			kubeClient:        kubeClient,
		}

		t.Logf(tc.name)
		err := bal.Add(ctx, &ba)
		if err != nil {
			t.Errorf("Add returned: %+v", err)
		}

		updatedBA, _ := bal.bucketAccesses(ns).Get(ctx, ba.ObjectMeta.Name, metav1.GetOptions{})
		if updatedBA.Status.AccessGranted != true {
			t.Errorf("Expected %t, got %t", true, ba.Status.AccessGranted)
		}
		if !strings.EqualFold(updatedBA.Status.AccountID, accountId) {
			t.Errorf("Expected %s, got %s", accountId, updatedBA.Status.AccountID)
		}

		_, err = bal.secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Secret creation failed: %v", err)
		}
	}
}
