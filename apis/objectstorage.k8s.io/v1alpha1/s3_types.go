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

// +k8s:deepcopy-gen=false

package v1alpha1

import (
	osspec "sigs.k8s.io/container-object-storage-interface-spec"
)

type S3SignatureVersion string

const (
	S3SignatureVersionV2 S3SignatureVersion = "S3V2"
	S3SignatureVersionV4 S3SignatureVersion = "S3V4"
)

type S3Protocol struct {
	Region     string `json:"region,omitempty"`
	// +kubebuilder:validation:Enum:={S3V2,S3V4}
	SignatureVersion S3SignatureVersion `json:"signatureVersion,omitempty"`
}

func (s3 *S3Protocol) ConvertToExternal() *osspec.Protocol_S3 {
	sigver, ok := osspec.S3SignatureVersion_value[string(s3.SignatureVersion)]
	if !ok {
		// NOTE - 0 here is equivalent to UnknownSignature
		sigver = 0
	}
	return &osspec.Protocol_S3{
		S3: &osspec.S3{
			Region:           s3.Region,
			SignatureVersion: osspec.S3SignatureVersion(sigver),
		},
	}
}

func ConvertFromS3External(ext *osspec.S3) *S3Protocol {
	vers, ok := osspec.S3SignatureVersion_name[int32(ext.SignatureVersion)]
	if !ok {
		vers = osspec.S3SignatureVersion_name[0]
	}
	return &S3Protocol{
		Region: ext.Region,
		SignatureVersion: S3SignatureVersion(vers),
	}
}
