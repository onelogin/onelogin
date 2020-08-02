package tfimport

import (
	"github.com/onelogin/onelogin/terraform/importables"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

type MockFile struct {
	Content []byte
}

func (m *MockFile) Write(p []byte) (int, error) {
	m.Content = p
	return len(p), nil
}

func (m *MockFile) Read(p []byte) (int, error) {
	for i, b := range m.Content {
		p[i] = b
	}
	return len(p), io.EOF
}

type MockImportable struct {
	InputResourceDefinitions []tfimportables.ResourceDefinition
}

func (mi MockImportable) ImportFromRemote() []tfimportables.ResourceDefinition {
	return mi.InputResourceDefinitions
}

func (mi MockImportable) HCLShape() interface{} {
	return nil
}

func TestFilterExistingDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputReadWriter             io.Reader
		Importable                  MockImportable
		ExpectedResourceDefinitions []tfimportables.ResourceDefinition
		ExpectedProviders           []string
	}{
		"it yields lists of resource definitions and providers not already defined in main.tf": {
			InputReadWriter: strings.NewReader(`
				resource onelogin_apps defined_in_main_already {
					name = defined_in_main_already
				}
				provider onelogin {
					alias "onelogin"
				}
			`),
			Importable: MockImportable{InputResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "defined_in_main_already", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "new_resource", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "test", Type: "onelogin_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "okra", Name: "test", Type: "okra_saml_apps"},
			}},
			ExpectedResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "new_resource", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "test", Type: "onelogin_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "okra", Name: "test", Type: "okra_saml_apps"},
			},
			ExpectedProviders: []string{"okra"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualResourceDefinitions, actualProviderDefinitions := FilterExistingDefinitions(test.InputReadWriter, test.Importable)
			assert.Equal(t, test.ExpectedResourceDefinitions, actualResourceDefinitions)
			assert.Equal(t, test.ExpectedProviders, actualProviderDefinitions)
		})
	}
}

func TestAppendDefinitionsToMainTF(t *testing.T) {
	tests := map[string]struct {
		TestFile                 MockFile
		InputResourceDefinitions []tfimportables.ResourceDefinition
		InputProviderDefinitions []string
		ExpectedOut              []byte
	}{
		"it adds provider and resource to the writer": {
			InputResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Name: "test", Type: "test", Provider: "test"},
				tfimportables.ResourceDefinition{Name: "test", Type: "test", Provider: "test2"},
			},
			TestFile:                 MockFile{},
			InputProviderDefinitions: []string{"test", "test2"},
			ExpectedOut:              []byte("provider test {\n\talias = \"test\"\n}\n\nprovider test2 {\n\talias = \"test2\"\n}\n\nresource test test {}\nresource test test {}\n"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := make([]byte, len(test.ExpectedOut))
			WriteHCLDefinitionHeaders(test.InputResourceDefinitions, test.InputProviderDefinitions, &test.TestFile)
			test.TestFile.Read(actual)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}
