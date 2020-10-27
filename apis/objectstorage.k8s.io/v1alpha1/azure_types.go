// +k8s:deepcopy-gen=false

package v1alpha1

type AzureProtocol struct {
	ContainerName  string `json:"containerName,omitempty"`
	StorageAccount string `json:"storageAccount,omitempty"`
}
