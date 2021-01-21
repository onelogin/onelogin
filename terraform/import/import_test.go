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

func TestFilterExistingDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputReadWriter             io.Reader
		IncomingResourceDefinitions []tfimportables.ResourceDefinition
		ExpectedResourceDefinitions []tfimportables.ResourceDefinition
		ExpectedProviders           []string
	}{
		"it yields lists of resource definitions and providers not already defined in main.tf": {
			InputReadWriter: strings.NewReader(`
				terraform {
				  required_providers {
				    onelogin = {
				      source = "onelogin/onelogin"
				    }
						okra = {
				      source = "okra/okra"
				    }
				  }
				}
				resource onelogin_apps defined_in_main_already {
					name = defined_in_main_already
				}
				resource okra_saml_apps test_defined_already {
					name = test_defined_already
				}
			`),
			IncomingResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Provider: "onelogin/onelogin", Name: "defined_in_main_already", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "okra/okra", Name: "test_defined_already", Type: "okra_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin/onelogin", Name: "new_resource", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin/onelogin", Name: "test", Type: "onelogin_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "okra/okra", Name: "test", Type: "okra_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "aws/aws", Name: "test", Type: "aws_apps"},
			},
			ExpectedResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Provider: "onelogin/onelogin", Name: "new_resource", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin/onelogin", Name: "test", Type: "onelogin_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "okra/okra", Name: "test", Type: "okra_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "aws/aws", Name: "test", Type: "aws_apps"},
			},
			ExpectedProviders: []string{"onelogin/onelogin", "okra/okra", "aws/aws"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualResourceDefinitions, actualProviderDefinitions := FilterExistingDefinitions(test.InputReadWriter, test.IncomingResourceDefinitions)
			assert.Equal(t, test.ExpectedResourceDefinitions, actualResourceDefinitions)
			assert.Equal(t, len(test.ExpectedProviders), len(actualProviderDefinitions))
		})
	}
}

func TestAddNewProvidersAndResourceHCL(t *testing.T) {
	tests := map[string]struct {
		TestFile                 MockFile
		InputResourceDefinitions []tfimportables.ResourceDefinition
		InputProviderDefinitions []string
		ExpectedOut              string
	}{
		"it adds provider and resource to the writer": {
			InputResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Name: "test", Type: "test", ImportID: "test", Provider: "test/test"},
				tfimportables.ResourceDefinition{Name: "test", Type: "test", ImportID: "test", Provider: "test2/test2"},
			},
			TestFile:                 MockFile{},
			InputProviderDefinitions: []string{"test/test", "test2/test2"},
			ExpectedOut:              "terraform {\n\trequired_providers {\n\t\ttest = {\n\t\t\tsource = \"test/test\"\n\t\t}\n\t\ttest2 = {\n\t\t\tsource = \"test2/test2\"\n\t\t}\n\t}\n}\nresource test test {}\nresource test test {}\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := AddNewProvidersAndResourceHCL(&test.TestFile, test.InputResourceDefinitions, test.InputProviderDefinitions)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}
