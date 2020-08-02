package stateparser

import (
	"fmt"
	"github.com/onelogin/onelogin/terraform/importables"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
									"name":          "test",
									"connector_id":  22,
									"rules":         []map[string]interface{}{{"actions": []map[string]interface{}{{"value": []string{"member_of", "asdf"}}}}},
									"provisioning":  map[string]bool{"enabled": true},
									"configuration": map[string]string{"provider_arn": "arn", "signature_algorithm": "sha-256"},
								},
							},
						},
					},
				},
			},
			ExpectedOutput: fmt.Sprintf("provider onelogin {\n\talias = \"onelogin\"\n}\n\nresource onelogin_apps test_resource {\n\tprovider = onelogin\n\tconnector_id = 22\n\tname = \"test\"\n\n\tconfiguration = {\n\t\tprovider_arn = \"arn\"\n\t\tsignature_algorithm = \"sha-256\"\n\t}\n\n\tprovisioning = {\n\t\tenabled = true\n\t}\n\n\trules {\n\n\t\tactions {\n\t\t\tvalue = [\"member_of\",\"asdf\"]\n\t\t}\n\t}\n}\n\n"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := ConvertTFStateToHCL(test.InputState, tfimportables.OneloginAppsImportable{})
			assert.Equal(t, len(test.ExpectedOutput), len(string(actual)))
		})
	}
}
