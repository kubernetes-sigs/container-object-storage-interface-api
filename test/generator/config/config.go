/*
Copyright 2024 The Kubernetes Authors.

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

package config

import (
	"bytes"
	"io"
	"os"

	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	"sigs.k8s.io/yaml"
)

// Load reads the YAML configuration from the provided file path and unmarshals it into a Config struct.
// It returns a pointer to the Config struct and any error encountered during the process.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, f)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	err = yaml.Unmarshal(buf.Bytes(), cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Config represents the configuration for generating tests.
// It includes various fields like driver details, deletion policies, authentication types, protocols,
// and parameters for BucketClass and BucketAccessClass.
type Config struct {
	// Templates contains paths to the YAML templates for various resources.
	Templates Templates `json:"templates"`

	// Driver is the name of the driver used in the test configuration.
	Driver string `json:"driver"`

	// DeletionPolicy is a list of deletion policies to be applied in the tests.
	DeletionPolicy []v1alpha1.DeletionPolicy `json:"deletionPolicy"`

	// AuthenticationType is a list of authentication types to be used in the tests.
	AuthenticationType []v1alpha1.AuthenticationType `json:"authenticationType"`

	// Protocol is a list of protocols to be used in the tests.
	Protocol []v1alpha1.Protocol `json:"protocol"`

	// BucketClassParams is a list of parameters for configuring BucketClass resources.
	BucketClassParams []map[string]string `json:"bucketClassParams"`

	// BucketAccessClassParams is a list of parameters for configuring BucketAccessClass resources.
	BucketAccessClassParams []map[string]string `json:"bucketAccessClassParams"`
}

// Templates represents the file paths to the templates used for generating various Kubernetes resources.
// These templates are referenced in the test configurations.
type Templates struct {
	// ChainsawTest is the path to the Chainsaw test YAML template.
	ChainsawTest string `json:"chainsaw-test.yaml"`

	// BucketAccess is the path to the BucketAccess YAML template.
	BucketAccess string `json:"BucketAccess.yaml"`

	// BucketAccessClass is the path to the BucketAccessClass YAML template.
	BucketAccessClass string `json:"BucketAccessClass.yaml"`

	// BucketClaim is the path to the BucketClaim YAML template.
	BucketClaim string `json:"BucketClaim.yaml"`

	// BucketClass is the path to the BucketClass YAML template.
	BucketClass string `json:"BucketClass.yaml"`
}
