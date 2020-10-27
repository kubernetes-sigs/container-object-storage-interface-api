package v1alpha1

type ProtocolName string

const (
	ProtocolNameS3    ProtocolName = "s3"
	ProtocolNameAzure ProtocolName = "azureBlob"
	ProtocolNameGCS   ProtocolName = "gcs"
)

type RequestedProtocol struct {
	// +kubebuilder:validation:Enum:={s3,azureBlob,gcs}
	Name ProtocolName `json:"name"`
	// +optional
	Version string `json:"version"`
}

type Protocol struct {
	// +required
	RequestedProtocol `json:"requestedProtocol"`
	// +optional
	S3 *S3Protocol `json:"s3,omitempty"`
	// +optional
	AzureBlob *AzureProtocol `json:"azureBlob,omitempty"`
	// +optional
	GCS *GCSProtocol `json:"gcs,omitempty"`
}
