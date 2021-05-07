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

import osspec "sigs.k8s.io/container-object-storage-interface-spec"

type AzureProtocol struct {
	StorageAccount string `json:"storageAccount,omitempty"`
}

func (azure *AzureProtocol) ConvertToExternal() *osspec.Protocol_AzureBlob {
	return &osspec.Protocol_AzureBlob{
		AzureBlob: &osspec.AzureBlob{
			StorageAccount: azure.StorageAccount,
		},
	}
}

func ConvertFromAzureExternal(ext *osspec.AzureBlob) *AzureProtocol {
	return &AzureProtocol{
		StorageAccount: ext.StorageAccount,
	}
}
