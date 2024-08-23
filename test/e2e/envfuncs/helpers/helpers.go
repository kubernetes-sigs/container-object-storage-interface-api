package helpers

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

type (
	NamespaceCtxKey string
)

// GetNamespaceKey returns the context key for a given test
func GetNamespaceKey(t *testing.T) NamespaceCtxKey {
	// When we pass t.Name() from inside an `assess` step, the name is in the form TestName/Features/Assess
	if strings.Contains(t.Name(), "/") {
		return NamespaceCtxKey(strings.Split(t.Name(), "/")[0])
	}

	// When pass t.Name() from inside a `testenv.BeforeEachTest` function, the name is just TestName
	return NamespaceCtxKey(t.Name())
}

func Load(path string) (*apiextensions.CustomResourceDefinition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)

	io.Copy(buf, f)

	crd := &apiextensions.CustomResourceDefinition{}

	if err := yaml.NewDecoder(buf).Decode(crd); err != nil {
		return nil, err
	}

	return crd, nil
}
