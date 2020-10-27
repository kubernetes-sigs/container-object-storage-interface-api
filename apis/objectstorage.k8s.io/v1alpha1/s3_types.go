// +k8s:deepcopy-gen=false

package v1alpha1

type S3SignatureVersion string

const (
	S3SignatureVersionV2 = "s3v2"
	S3SignatureVersionV4 = "s3v4"
)

type S3Protocol struct {
	Version    string `json:"version,omitempty"`
	Endpoint   string `json:"endpoint,omitempty"`
	BucketName string `json:"bucketName,omitempty"`
	Region     string `json:"region,omitempty"`
	// +kubebuilder:validation:Enum:={s3v2,s3v4}
	SignatureVersion S3SignatureVersion `json:"signatureVersion,omitempty"`
}
