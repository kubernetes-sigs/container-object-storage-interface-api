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

type S3SignatureVersion string

const (
	S3SignatureVersionV2 S3SignatureVersion = "S3v2"
	S3SignatureVersionV4 S3SignatureVersion = "S3v4"
)

type S3Protocol struct {
	Endpoint   string `json:"endpoint,omitempty"`
	BucketName string `json:"bucketName,omitempty"`
	Region     string `json:"region,omitempty"`
	// +kubebuilder:validation:Enum:={s3v2,s3v4}
	SignatureVersion S3SignatureVersion `json:"signatureVersion,omitempty"`
}
