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

	osspec "sigs.k8s.io/container-object-storage-interface-spec"
)

type ProtocolName string

const (
	ProtocolNameS3    ProtocolName = "s3"
	ProtocolNameAzure ProtocolName = "azureBlob"
	ProtocolNameGCS   ProtocolName = "gcs"

	MissingS3Protocol = "missing s3 in protocol"
	MissingAzureProtocol = "missing azure in protocol"
	MissingGCSProtocol = "missing gcs in protocol"
	InvalidProtocolName = "invalid protocol name"
)

type Protocol struct {
	// +kubebuilder:validation:Enum:={s3,azureBlob,gcs}
	Name ProtocolName `json:"name"`
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	S3 *S3Protocol `json:"s3,omitempty"`
	// +optional
	AzureBlob *AzureProtocol `json:"azureBlob,omitempty"`
	// +optional
	GCS *GCSProtocol `json:"gcs,omitempty"`
}

func (in *Protocol) ConvertToExternal() (*osspec.Protocol, error) {
	external := &osspec.Protocol{
		Version: in.Version,
	}

	switch in.Name {
	case ProtocolNameS3:
		if in.S3 == nil {
			return nil, errors.New(MissingS3Protocol)
		}
		external.Name = osspec.ProtocolName_S3
		external.Type = in.S3.ConvertToExternal()
	case ProtocolNameAzure:
		if in.AzureBlob == nil {
			return nil, errors.New(MissingAzureProtocol)
		}
		external.Name = osspec.ProtocolName_AZURE
		external.Type = in.AzureBlob.ConvertToExternal()
	case ProtocolNameGCS:
		if in.GCS == nil {
			return nil, errors.New(MissingGCSProtocol)
		}
		external.Name = osspec.ProtocolName_GCS
		external.Type = in.GCS.ConvertToExternal()
	default:
		external.Name = osspec.ProtocolName_UnknownProtocol
		return external, errors.New(InvalidProtocolName)

	}

	return external, nil
}

func ConvertFromProtocolExternal(external *osspec.Protocol) (*Protocol, error) {
	in := &Protocol{}
	in.Version = external.Version

	switch external.Name {
	case osspec.ProtocolName_S3:
		if external.GetS3() == nil {
			return nil, errors.New(MissingS3Protocol)
		}
		in.Name = ProtocolNameS3
		in.S3 = ConvertFromS3External(external.GetS3())
	case osspec.ProtocolName_AZURE:
		if external.GetAzureBlob() == nil {
			return nil, errors.New(MissingAzureProtocol)
		}
		in.Name = ProtocolNameAzure
		in.AzureBlob = ConvertFromAzureExternal(external.GetAzureBlob())
	case osspec.ProtocolName_GCS:
		if external.GetGcs() == nil {
			return nil, errors.New(MissingGCSProtocol)
		}
		in.Name = ProtocolNameGCS
		in.GCS = ConvertFromGCSExternal(external.GetGcs())
	default:
		// TODO - Do we to set the protocol Name to specific value here?
		return nil, errors.New(InvalidProtocolName)
	}

	return in, nil
}
