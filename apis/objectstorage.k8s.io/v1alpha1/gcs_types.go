// +k8s:deepcopy-gen=false

package v1alpha1

type GCSProtocol struct {
	BucketName     string `json:"bucketName,omitempty"`
	PrivateKeyName string `json:"privateKeyName,omitempty"`
	ProjectID      string `json:"projectID,omitempty"`
	ServiceAccount string `json:"serviceAccount,omitempty"`
}
