/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bucket

import (
	"context"
	"reflect"
	"testing"

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage.k8s.io/v1alpha1"

	fakebucketclientset "sigs.k8s.io/container-object-storage-interface-api/clientset/fake"

	osspec "sigs.k8s.io/container-object-storage-interface-spec"
	fakespec "sigs.k8s.io/container-object-storage-interface-spec/fake"

	corev1 "k8s.io/api/core/v1"

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
		t.Fatalf("couldn't convert Discovery() to *FakeDiscovery")
	}

	fakeVersion := &version.Info{
		GitVersion: "v1.0.0",
	}
	fakeDiscovery.FakedServerVersion = fakeVersion

	bl := bucketListener{}
	bl.InitializeKubeClient(client)

	if bl.kubeClient == nil {
		t.Errorf("kubeClient was nil")
	}

	expected := utilversion.MustParseSemantic(fakeVersion.GitVersion)
	if !reflect.DeepEqual(expected, bl.kubeVersion) {
		t.Errorf("expected %+v, but got %+v", expected, bl.kubeVersion)
	}
}

func TestInitializeBucketClient(t *testing.T) {
	client := fakebucketclientset.NewSimpleClientset()

	bl := bucketListener{}
	bl.InitializeBucketClient(client)

	if bl.bucketClient == nil {
		t.Errorf("bucketClient was nil")
	}
}

func TestAddWrongProvisioner(t *testing.T) {
	provisioner := "provisioner1"
	mpc := struct{ fakespec.FakeProvisionerClient }{}
	mpc.FakeProvisionerCreateBucket = func(ctx context.Context, in *osspec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerCreateBucketResponse, error) {
		t.Errorf("grpc client called")
		return nil, nil
	}

	bl := bucketListener{
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
		t.Errorf("error returned: %+v", err)
	}
}

func TestAddValidProtocols(t *testing.T) {
	provisioner := "provisioner1"
	protocolVersion := "proto1"
	anonAccess := "BUCKET_PRIVATE"
	bucketName := "bucket1"
	s3 := v1alpha1.S3Protocol{
		BucketName:       "bucket1",
		Endpoint:         "127.0.0.1",
		Region:           "region1",
		SignatureVersion: v1alpha1.S3SignatureVersionV2,
	}
	gcs := v1alpha1.GCSProtocol{
		BucketName:     "bucket1",
		PrivateKeyName: "keyName1",
		ProjectID:      "id1",
		ServiceAccount: "account1",
	}
	azure := v1alpha1.AzureProtocol{
		ContainerName:  "bucket1",
		StorageAccount: "account1",
	}
	mpc := struct{ fakespec.FakeProvisionerClient }{}

	testCases := []struct {
		name         string
		protocolName v1alpha1.ProtocolName
		setProtocol  func(b *v1alpha1.Bucket)
		createFunc   func(ctx context.Context, in *osspec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerCreateBucketResponse, error)
		params       map[string]string
	}{
		{
			name:         "S3",
			protocolName: v1alpha1.ProtocolNameS3,
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.S3 = &s3
			},
			createFunc: func(ctx context.Context, req *osspec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerCreateBucketResponse, error) {
				in := req.Protocol.GetS3()
				if in.BucketName != s3.BucketName {
					t.Errorf("expected %s, got %s", s3.BucketName, in.BucketName)
				}
				if in.Region != s3.Region {
					t.Errorf("expected %s, got %s", s3.Region, in.Region)
				}
				sigver, ok := osspec.S3SignatureVersion_name[int32(in.SignatureVersion)]
				if !ok {
					sigver = osspec.S3SignatureVersion_name[int32(osspec.S3SignatureVersion_UnknownSignature)]
				}
				if sigver != string(s3.SignatureVersion) {
					t.Errorf("expected %s, got %s", s3.SignatureVersion, sigver)
				}
				if in.Endpoint != s3.Endpoint {
					t.Errorf("expected %s, got %s", in.Endpoint, in.Endpoint)
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerCreateBucketResponse{}, nil
			},
		},
		{
			name:         "GCS",
			protocolName: v1alpha1.ProtocolNameGCS,
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.GCS = &gcs
			},
			createFunc: func(ctx context.Context, req *osspec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerCreateBucketResponse, error) {
				in := req.Protocol.GetGcs()
				if in.BucketName != gcs.BucketName {
					t.Errorf("expected %s, got %s", gcs.BucketName, in.BucketName)
				}
				if in.ServiceAccount != gcs.ServiceAccount {
					t.Errorf("expected %s, got %s", gcs.ServiceAccount, in.ServiceAccount)
				}
				if in.PrivateKeyName != gcs.PrivateKeyName {
					t.Errorf("expected %s, got %s", gcs.PrivateKeyName, in.PrivateKeyName)
				}
				if in.ProjectId != gcs.ProjectID {
					t.Errorf("expected %s, got %s", gcs.ProjectID, in.ProjectId)
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerCreateBucketResponse{}, nil
			},
		},
		{
			name:         "AzureBlob",
			protocolName: v1alpha1.ProtocolNameAzure,
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.AzureBlob = &azure
			},
			createFunc: func(ctx context.Context, req *osspec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerCreateBucketResponse, error) {
				in := req.Protocol.GetAzureBlob()
				if in.ContainerName != azure.ContainerName {
					t.Errorf("expected %s, got %s", azure.ContainerName, in.ContainerName)
				}
				if in.StorageAccount != azure.StorageAccount {
					t.Errorf("expected %s, got %s", azure.StorageAccount, in.StorageAccount)
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerCreateBucketResponse{}, nil
			},
		},
		{
			name:         "AnonymousAccessMode",
			protocolName: v1alpha1.ProtocolNameAzure,
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.AzureBlob = &azure
			},
			createFunc: func(ctx context.Context, req *osspec.ProvisionerCreateBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerCreateBucketResponse, error) {
				in := req.Protocol.GetAzureBlob()
				if in.ContainerName != azure.ContainerName {
					t.Errorf("expected %s, got %s", azure.ContainerName, in.ContainerName)
				}
				if in.StorageAccount != azure.StorageAccount {
					t.Errorf("expected %s, got %s", azure.StorageAccount, in.StorageAccount)
				}
				aMode := osspec.AnonymousBucketAccessMode(osspec.AnonymousBucketAccessMode_value[anonAccess])
				if req.AnonymousBucketAccessMode != aMode {
					t.Errorf("expected %s, got %s", aMode, req.AnonymousBucketAccessMode)
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerCreateBucketResponse{}, nil
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
				Protocol: v1alpha1.Protocol{
					Name:    tc.protocolName,
					Version: protocolVersion,
				},
				Parameters: tc.params,
			},
		}

		ctx := context.TODO()
		client := fakebucketclientset.NewSimpleClientset(&b)
		kubeClient := fakekubeclientset.NewSimpleClientset()
		mpc.FakeProvisionerCreateBucket = tc.createFunc
		bl := bucketListener{
			provisionerName:   provisioner,
			provisionerClient: &mpc,
			bucketClient:      client,
			kubeClient:        kubeClient,
		}

		tc.setProtocol(&b)
		t.Logf(tc.name)
		err := bl.Add(ctx, &b)
		if err != nil {
			t.Errorf("add returned: %+v", err)
		}

		updatedB, _ := client.ObjectstorageV1alpha1().Buckets().Get(ctx, b.Name, metav1.GetOptions{})
		if updatedB.Status.BucketAvailable != true {
			t.Errorf("expected %t, got %t", true, b.Status.BucketAvailable)
		}
	}
}

func TestDeleteWrongProvisioner(t *testing.T) {
	provisioner := "provisioner1"
	mpc := struct{ fakespec.FakeProvisionerClient }{}
	mpc.FakeProvisionerDeleteBucket = func(ctx context.Context, in *osspec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerDeleteBucketResponse, error) {
		t.Errorf("grpc client called")
		return nil, nil
	}

	bl := bucketListener{
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

func TestDeleteValidProtocols(t *testing.T) {
	provisioner := "provisioner1"
	region := "region1"
	bucketName := "bucket1"
	protocolVersion := "proto1"
	sigVersion := v1alpha1.S3SignatureVersion(v1alpha1.S3SignatureVersionV2)
	account := "account1"
	keyName := "keyName1"
	projID := "id1"
	endpoint := "endpoint1"
	mpc := struct{ fakespec.FakeProvisionerClient }{}
	extraParamName := "ParamName"
	extraParamValue := "ParamValue"

	testCases := []struct {
		name         string
		setProtocol  func(b *v1alpha1.Bucket)
		protocolName v1alpha1.ProtocolName
		deleteFunc   func(ctx context.Context, in *osspec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerDeleteBucketResponse, error)
		params       map[string]string
	}{
		{
			name: "S3",
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.S3 = &v1alpha1.S3Protocol{
					Region:           region,
					SignatureVersion: sigVersion,
					BucketName:       bucketName,
					Endpoint:         endpoint,
				}
			},
			protocolName: v1alpha1.ProtocolNameS3,
			deleteFunc: func(ctx context.Context, req *osspec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerDeleteBucketResponse, error) {
				in := req.Protocol.GetS3()
				if in.BucketName != bucketName {
					t.Errorf("expected %s, got %s", bucketName, in.BucketName)
				}
				if in.Region != region {
					t.Errorf("expected %s, got %s", region, in.Region)
				}
				sigver, ok := osspec.S3SignatureVersion_name[int32(in.SignatureVersion)]
				if !ok {
					sigver = osspec.S3SignatureVersion_name[int32(osspec.S3SignatureVersion_UnknownSignature)]
				}
				if sigver != string(sigVersion) {
					t.Errorf("expected %s, got %s", sigVersion, sigver)
				}
				if in.Endpoint != endpoint {
					t.Errorf("expected %s, got %s", endpoint, in.Endpoint)
				}
				if req.Parameters[extraParamName] != extraParamValue {
					t.Errorf("expected %s, got %s", extraParamValue, req.Parameters[extraParamName])
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerDeleteBucketResponse{}, nil
			},
			params: map[string]string{
				extraParamName: extraParamValue,
			},
		},
		{
			name: "GCS",
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.GCS = &v1alpha1.GCSProtocol{
					ServiceAccount: account,
					PrivateKeyName: keyName,
					ProjectID:      projID,
					BucketName:     bucketName,
				}
			},
			protocolName: v1alpha1.ProtocolNameGCS,
			deleteFunc: func(ctx context.Context, req *osspec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerDeleteBucketResponse, error) {
				in := req.Protocol.GetGcs()
				if in.BucketName != bucketName {
					t.Errorf("expected %s, got %s", bucketName, in.BucketName)
				}
				if in.ServiceAccount != account {
					t.Errorf("expected %s, got %s", region, in.ServiceAccount)
				}
				if in.PrivateKeyName != keyName {
					t.Errorf("expected %s, got %s", region, in.PrivateKeyName)
				}
				if in.ProjectId != projID {
					t.Errorf("expected %s, got %s", region, in.ProjectId)
				}
				if req.Parameters[extraParamName] != extraParamValue {
					t.Errorf("expected %s, got %s", extraParamValue, req.Parameters[extraParamName])
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerDeleteBucketResponse{}, nil
			},
			params: map[string]string{
				extraParamName: extraParamValue,
			},
		},
		{
			name: "AzureBlob",
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.AzureBlob = &v1alpha1.AzureProtocol{
					StorageAccount: account,
					ContainerName:  bucketName,
				}
			},
			protocolName: v1alpha1.ProtocolNameAzure,
			deleteFunc: func(ctx context.Context, req *osspec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerDeleteBucketResponse, error) {
				in := req.Protocol.GetAzureBlob()
				if in.ContainerName != bucketName {
					t.Errorf("expected %s, got %s", bucketName, in.ContainerName)
				}
				if in.StorageAccount != account {
					t.Errorf("expected %s, got %s", region, in.StorageAccount)
				}
				if req.Parameters[extraParamName] != extraParamValue {
					t.Errorf("expected %s, got %s", extraParamValue, req.Parameters[extraParamName])
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerDeleteBucketResponse{}, nil
			},
			params: map[string]string{
				extraParamName: extraParamValue,
			},
		},
		{
			name: "Empty parameters",
			setProtocol: func(b *v1alpha1.Bucket) {
				b.Spec.Protocol.S3 = &v1alpha1.S3Protocol{
					Region:           region,
					SignatureVersion: sigVersion,
					BucketName:       bucketName,
					Endpoint:         endpoint,
				}
			},
			protocolName: v1alpha1.ProtocolNameS3,
			deleteFunc: func(ctx context.Context, req *osspec.ProvisionerDeleteBucketRequest, opts ...grpc.CallOption) (*osspec.ProvisionerDeleteBucketResponse, error) {
				in := req.Protocol.GetS3()
				if in.BucketName != bucketName {
					t.Errorf("expected %s, got %s", bucketName, in.BucketName)
				}
				if in.Region != region {
					t.Errorf("expected %s, got %s", region, in.Region)
				}
				sigver, ok := osspec.S3SignatureVersion_name[int32(in.SignatureVersion)]
				if !ok {
					sigver = osspec.S3SignatureVersion_name[int32(osspec.S3SignatureVersion_UnknownSignature)]
				}
				if sigver != string(sigVersion) {
					t.Errorf("expected %s, got %s", sigVersion, sigver)
				}
				if in.Endpoint != endpoint {
					t.Errorf("expected %s, got %s", endpoint, in.Endpoint)
				}
				if req.Parameters["ProtocolVersion"] != protocolVersion {
					t.Errorf("expected %s, got %s", protocolVersion, req.Parameters["ProtocolVersion"])
				}
				return &osspec.ProvisionerDeleteBucketResponse{}, nil
			},
			params: nil,
		},
	}

	for _, tc := range testCases {
		b := v1alpha1.Bucket{
			Spec: v1alpha1.BucketSpec{
				Provisioner: provisioner,
				Protocol: v1alpha1.Protocol{
					Name:    tc.protocolName,
					Version: protocolVersion,
				},
				Parameters: tc.params,
			},
			Status: v1alpha1.BucketStatus{
				BucketAvailable: true,
			},
		}

		ctx := context.TODO()
		client := fakebucketclientset.NewSimpleClientset(&b)
		mpc.FakeProvisionerDeleteBucket = tc.deleteFunc
		bl := bucketListener{
			provisionerName:   provisioner,
			provisionerClient: &mpc,
			bucketClient:      client,
		}

		tc.setProtocol(&b)
		t.Logf(tc.name)
		err := bl.Delete(ctx, &b)
		if err != nil {
			t.Errorf("delete returned: %+v", err)
		}

		updatedB, _ := client.ObjectstorageV1alpha1().Buckets().Get(ctx, b.Name, metav1.GetOptions{})
		if updatedB.Status.BucketAvailable != false {
			t.Errorf("expected %t, got %t", false, b.Status.BucketAvailable)
		}
	}
}

func TestDeleteInvalidProtocol(t *testing.T) {
	const (
		protocolName v1alpha1.ProtocolName = "invalid"
	)

	bucketName := "bucket1"
	provisioner := "provisioner1"

	bl := bucketListener{
		provisionerName: provisioner,
	}

	b := v1alpha1.Bucket{
		Spec: v1alpha1.BucketSpec{
			BucketRequest: &corev1.ObjectReference{
				Name: bucketName,
			},
			Provisioner: provisioner,
			Protocol: v1alpha1.Protocol{
				Name: protocolName,
			},
		},
	}

	ctx := context.TODO()
	err := bl.Delete(ctx, &b)
	if err == nil {
		t.Errorf("invalidProtocol: no error returned")
	}
}
