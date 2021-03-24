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

package v1alpha1

import (
	"errors"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

type ProtocolName string

const (
	ProtocolNameS3    ProtocolName = "s3"
	ProtocolNameAzure ProtocolName = "azureBlob"
	ProtocolNameGCS   ProtocolName = "gcs"

	InvalidProtocol = "invalid protocol"
)

type Protocol struct {
	// +optional
	S3 *S3Protocol `json:"s3,omitempty"`

	// +optional
	AzureBlob *AzureProtocol `json:"azureBlob,omitempty"`

	// +optional
	GCS *GCSProtocol `json:"gcs,omitempty"`
}

func (in *Protocol) ConvertToExternal() (*cosi.Protocol, error) {
	external := &cosi.Protocol{}

	protoFound := false
	if in.S3 != nil {
		protoFound = true
		external.Type = in.S3.ConvertToExternal()
	}
	if in.AzureBlob != nil {
		protoFound = true
		external.Type = in.AzureBlob.ConvertToExternal()
	}
	if in.GCS != nil {
		protoFound = true
		external.Type = in.GCS.ConvertToExternal()
	}

	if !protoFound {
		return external, errors.New(InvalidProtocol)
	}

	return external, nil
}

func ConvertFromProtocolExternal(external *cosi.Protocol) (*Protocol, error) {
	in := &Protocol{}

	protoFound := false
	if external.GetS3() != nil {
		protoFound = true
		in.S3 = ConvertFromS3External(external.GetS3())
	}
	if external.GetAzureBlob() != nil {
		protoFound = true
		in.AzureBlob = ConvertFromAzureExternal(external.GetAzureBlob())
	}
	if external.GetGcs() != nil {
		protoFound = true
		in.GCS = ConvertFromGCSExternal(external.GetGcs())
	}
	if !protoFound {
		return nil, errors.New(InvalidProtocol)
	}

	return in, nil
}
