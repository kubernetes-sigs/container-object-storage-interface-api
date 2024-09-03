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

package generator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	// embedding Chainsaw test templates
	_ "embed"

	"k8s.io/klog/v2"
	"sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	"sigs.k8s.io/container-object-storage-interface-api/test/generator/config"
)

var (
	// chainsawTestSpec is the embedded template for the Chainsaw test specification.
	//go:embed template/chainsaw-test.gotpl
	chainsawTestSpec string

	// bucketClassTpl is the embedded template for the BucketClass resource.
	//go:embed template/BucketClass.gotpl
	bucketClassTpl string

	// bucketAccessClassTpl is the embedded template for the BucketAccessClass resource.
	//go:embed template/BucketAccessClass.gotpl
	bucketAccessClassTpl string

	// bucketClaimTpl is the embedded template for the BucketClaim resource.
	//go:embed template/BucketClaim.gotpl
	bucketClaimTpl string

	// bucketAccessTpl is the embedded template for the BucketAccess resource.
	//go:embed template/BucketAccess.gotpl
	bucketAccessTpl string
)

// Test represents a test case that includes a name and associated resources.
type Test struct {
	Name             string
	Resources        Resources
	ChainsawTestSpec []byte
}

// Resources contains the byte representations of various Kubernetes resources used in a test.
type Resources struct {
	BucketClass       []byte
	BucketAccessClass []byte
	BucketClaim       []byte
	BucketAccess      []byte
}

const (
	Name      = "name"
	Driver    = "driver"
	Proto     = "protocol"
	Auth      = "authenticationType"
	Policy    = "deletionPolicy"
	BcParams  = "bucketClassParams"
	BacParams = "bucketAccessClassParams"
)

// matrix represents a test matrix, which is a collection of test data variations.
type matrix map[string]map[string]any

// newMatrix creates a new matrix with a given driver.
func newMatrix(driver string) matrix {
	return matrix{
		"test": map[string]any{
			Driver: driver,
			Name:   "test",
		},
	}
}

// addProtocol adds protocol variations to the matrix.
func (src matrix) addProtocol(protocols []v1alpha1.Protocol) matrix {
	override := matrix{}

	if len(protocols) == 0 {
		return src
	}

	for _, p := range protocols {
		for name, data := range src {
			name = fmt.Sprintf(
				"%s-%s",
				name,
				strings.ToLower(string(p)),
			)

			data[Proto] = string(p)
			data[Name] = name

			override[name] = data
		}
	}

	return override
}

// addAuth adds authentication type variations to the matrix.
func (src matrix) addAuth(auths []v1alpha1.AuthenticationType) matrix {
	override := matrix{}

	if len(auths) == 0 {
		return src
	}

	for _, at := range auths {
		for name, data := range src {
			name := fmt.Sprintf(
				"%s-%s",
				name,
				strings.ToLower(string(at)),
			)

			data[Auth] = string(at)
			data[Name] = name

			override[name] = data
		}
	}

	return override
}

// addDeletionPolicy adds deletion policy variations to the matrix.
func (src matrix) addDeletionPolicy(policies []v1alpha1.DeletionPolicy) matrix {
	override := matrix{}

	if len(policies) == 0 {
		return src
	}

	for _, dp := range policies {
		for name, data := range src {
			name = fmt.Sprintf(
				"%s-%s",
				name,
				strings.ToLower(string(dp)),
			)

			data[Policy] = string(dp)
			data[Name] = name

			override[name] = data
		}
	}

	return override
}

// addBCParams adds BucketClass parameters to the matrix.
func (src matrix) addBCParams(params []map[string]string) matrix {
	override := matrix{}

	if len(params) == 0 {
		return src
	}

	for bcp, params := range params {
		for name, data := range src {
			data := copy(data)
			sbcp := fmt.Sprintf("bcp%d", bcp)
			name := fmt.Sprintf(
				"%s-%s",
				name,
				sbcp,
			)

			data[BcParams] = copy(params)
			data[Name] = name

			override[name] = data
		}
	}

	return override
}

// addBACParams adds BucketAccessClass parameters to the matrix.
func (src matrix) addBACParams(params []map[string]string) matrix {
	override := matrix{}

	if len(params) == 0 {
		return src
	}

	for bacp, params := range params {
		for name, data := range src {
			data := copy(data)
			sbacp := fmt.Sprintf("bacp%d", bacp)
			name := fmt.Sprintf(
				"%s-%s",
				name,
				sbacp,
			)
			newParams := copy(params)

			data[BacParams] = newParams
			data[Name] = name

			override[name] = data
		}
	}

	return override
}

// toTests converts the matrix into a slice of Test objects using the provided templates.
func (src matrix) toTests(tpls config.Templates) ([]Test, error) {
	tests := []Test{}

	for name, data := range src {
		chainsawTest, err := renderTemplate(
			loadTemplateOrDefault(tpls.ChainsawTest, chainsawTestSpec),
			data,
		)
		if err != nil {
			return nil, err
		}

		bucketClass, err := renderTemplate(
			loadTemplateOrDefault(tpls.ChainsawTest, bucketClassTpl),
			data,
		)
		if err != nil {
			return nil, err
		}

		bucketAccessClass, err := renderTemplate(
			loadTemplateOrDefault(tpls.ChainsawTest, bucketAccessClassTpl),
			data,
		)
		if err != nil {
			return nil, err
		}

		bucketAccess, err := renderTemplate(
			loadTemplateOrDefault(tpls.ChainsawTest, bucketAccessTpl),
			data,
		)
		if err != nil {
			return nil, err
		}

		bucketClaim, err := renderTemplate(
			loadTemplateOrDefault(tpls.ChainsawTest, bucketClaimTpl),
			data,
		)
		if err != nil {
			return nil, err
		}

		tests = append(tests, Test{
			Name:             name,
			ChainsawTestSpec: chainsawTest,
			Resources: Resources{
				BucketClass:       bucketClass,
				BucketAccessClass: bucketAccessClass,
				BucketAccess:      bucketAccess,
				BucketClaim:       bucketClaim,
			},
		})
	}

	return tests, nil
}

// Matrix generates a slice of tests based on the provided configuration.
func Matrix(cfg *config.Config) ([]Test, error) {
	return newMatrix(cfg.Driver).
		addAuth(cfg.AuthenticationType).
		addBACParams(cfg.BucketAccessClassParams).
		addBCParams(cfg.BucketClassParams).
		addDeletionPolicy(cfg.DeletionPolicy).
		addProtocol(cfg.Protocol).
		toTests(cfg.Templates)
}

// copy copies all key-value pairs from the src map to a new map.
func copy[T comparable, U any](src map[T]U) map[T]U {
	dst := make(map[T]U)

	// Iterate over the source map
	for key, value := range src {
		// Copy each key-value pair to the destination map
		dst[key] = value
	}

	return dst
}

// renderTemplate renders a template with the provided data and returns the result as a byte slice.
func renderTemplate(text string, data map[string]any) ([]byte, error) {
	buf := new(bytes.Buffer)

	tpl, err := template.New("").Parse(text)
	if err != nil {
		return nil, err
	}

	if err := tpl.Execute(buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// loadTemplateOrDefault loads a template from a file if a path is provided.
// If the path is empty or an error occurs, it returns the default template text.
func loadTemplateOrDefault(path string, text string) string {
	if path == "" {
		return text
	}

	f, err := os.Open(path)
	if err != nil {
		klog.V(0).ErrorS(err, "Unable to open file", "file", path)
		return text
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		klog.V(0).ErrorS(err, "Unable to read file contents", "file", path)
		return text
	}

	return buf.String()
}
