package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:storageversion

type BucketAccessInfo struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	S3    *S3BucketAccessInfo    `json:"s3,omitempty"`
	Azure *AzureBucketAccessInfo `json:"azure,omitempty"`
	Gcs   *GCSBucketAccessInfo   `json:"gcs,omitempty"`
}

type S3BucketAccessInfo struct {
	Endpoint         string             `json:"endpoint,omitempty"`
	BucketName       string             `json:"bucketName,omitempty"`
	Region           string             `json:"region,omitempty"`
	Credentials      string             `json:"credentials,omitempty"`
	Certificates     string             `json:"certificates,omitempty"`
	SignatureVersion S3SignatureVersion `json:"signatureVersion,omitempty"`
}

type AzureBucketAccessInfo struct {
	Endpoint           string `json:"endpoint,omitempty"`           // scheme is mandatory in the URL
	StorageAccountName string `json:"storageAccountName,omitempty"` // This is the equivalent of bucket name + access key
	ContainerName      string `json:"containerName,omitempty"`      // Optional. This is a prefix at the root of the bucket
	SecretKey          string `json:"secretKey,omitempty"`          // This is the equivalent of secret key
}

type GCSBucketAccessInfo struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BucketAccessInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketAccessInfo `json:"items"`
}
