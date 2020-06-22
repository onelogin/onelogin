package tfimport

import (
	"fmt"
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
	return len(p), nil
}

func TestFilterExistingDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputReadWriter             io.Reader
		InputResourceDefinitions    []tfimportables.ResourceDefinition
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
			InputResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "defined_in_main_already", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "new_resource", Type: "onelogin_apps"},
				tfimportables.ResourceDefinition{Provider: "onelogin", Name: "test", Type: "onelogin_saml_apps"},
				tfimportables.ResourceDefinition{Provider: "okra", Name: "test", Type: "okra_saml_apps"},
			},
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
			actualResourceDefinitions, actualProviderDefinitions := filterExistingDefinitions(test.InputReadWriter, test.InputResourceDefinitions)
			assert.Equal(t, test.ExpectedResourceDefinitions, actualResourceDefinitions)
			assert.Equal(t, test.ExpectedProviders, actualProviderDefinitions)
		})
	}
}

func TestConvertTFStateToHCL(t *testing.T) {
	tests := map[string]struct {
		InputState     State
		ExpectedOutput string
	}{
		"it creates a bytes buffer representing the state in HCL": {
			InputState: State{
				Resources: []StateResource{
					StateResource{
						Name:     "test_resource",
						Type:     "onelogin_apps",
						Provider: "provider.onelogin",
						Instances: []ResourceInstance{
							ResourceInstance{
								Data: map[string]interface{}{
									"name":         "test",
									"connector_id": 22,
									"rules":        []map[string]interface{}{{"actions": []map[string]interface{}{{"value": []string{"member_of", "asdf"}}}}},
									"provisioning": []map[string]bool{{"enabled": true}},
								},
							},
						},
					},
				},
			},
			ExpectedOutput: fmt.Sprintf("provider onelogin {\n\talias = \"onelogin\"\n}\n\nresource onelogin_apps test_resource {\n\tprovider = onelogin\n\n\trules {\n\n\t\tactions {\n\t\t\tvalue = [\"member_of\",\"asdf\"]\n\t\t}\n\t}\n\tconnector_id = 22\n\tname = \"test\"\n\n\tprovisioning {\n\t\tenabled = true\n\t}\n}\n\n"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := convertTFStateToHCL(test.InputState)
			assert.Equal(t, len(test.ExpectedOutput), len(string(actual)))
		})
	}
}

func TestAppendDefinitionsToMainTF(t *testing.T) {
	tests := map[string]struct {
		InputWriter              io.ReadWriter
		InputResourceDefinitions []tfimportables.ResourceDefinition
		InputProviderDefinitions []string
		ExpectedOut              []byte
	}{
		"it adds provider and resource to the writer": {
			InputWriter: &MockFile{},
			InputResourceDefinitions: []tfimportables.ResourceDefinition{
				tfimportables.ResourceDefinition{Name: "test", Type: "test", Provider: "test"},
				tfimportables.ResourceDefinition{Name: "test", Type: "test", Provider: "test2"},
			},
			InputProviderDefinitions: []string{"test", "test2"},
			ExpectedOut:              []byte("provider test {\n\talias = \"test\"\n}\n\nprovider test2 {\n\talias = \"test2\"\n}\n\nresource test test {}\nresource test test {}\n"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := make([]byte, len(test.ExpectedOut))
			actual = createHCLDefinitionsBuffer(test.InputResourceDefinitions, test.InputProviderDefinitions)
			test.InputWriter.Read(actual)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}
