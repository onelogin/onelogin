package stateparser

import (
	"fmt"

	"testing"

	"github.com/onelogin/onelogin/clients"
	tfimportables "github.com/onelogin/onelogin/terraform/importables"
	"github.com/stretchr/testify/assert"
)

func TestConvertTFStateToHCL(t *testing.T) {
	tests := map[string]struct {
		InputState     State
		ExpectedOutput string
	}{
		"it creates a bytes buffer representing the state in HCL": {
			InputState: State{
				Resources: []StateResource{
					{
						Name:     "test_resource",
						Type:     "onelogin_apps",
						Provider: "provider[\"registry.terraform.io/onelogin/onelogin\"]",
						Instances: []ResourceInstance{
							{
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
					{
						Name:     "test_resource",
						Type:     "onelogin_roles",
						Provider: "provider[\"registry.terraform.io/onelogin/onelogin\"]",
						Instances: []ResourceInstance{
							{
								Data: map[string]interface{}{
									"name": "test",
									"apps": []int{1, 2, 3},
								},
							},
						},
					},
					{
						Name:     "test_resource",
						Type:     "onelogin_users",
						Provider: "provider[\"registry.terraform.io/onelogin/onelogin\"]",
						Instances: []ResourceInstance{
							{
								Data: map[string]interface{}{
									"username": "test",
									"email":    "test@test.test",
								},
							},
						},
					},
					{
						Name:     "test_resource",
						Type:     "aws_iam_user",
						Provider: "provider[\"registry.terraform.io/aws/aws\"]",
						Instances: []ResourceInstance{
							{
								Data: map[string]interface{}{
									"username": "test",
									"path":     "/",
								},
							},
						},
					},
				},
			},
			ExpectedOutput: fmt.Sprintf("terraform {\n\trequired_providers {\n\t\tonelogin = {\n\t\t\tsource = \"onelogin/onelogin\"\n\t\t}\n\t\taws = {\n\t\t\tsource = \"aws/aws\"\n\t\t}\n\t}\n}\nresource onelogin_apps test_resource {\n\n\tprovisioning = {\n\t\tenabled = true\n\t}\n\n\trules {\n\n\t\tactions {\n\t\t\tvalue = [\"member_of\", \"asdf\"]\n\t\t}\n\t}\n\tconnector_id = 22\n\tname = \"test\"\n\n\tconfiguration = {\n\t\tprovider_arn = \"arn\"\n\t\tsignature_algorithm = \"sha-256\"\n\t}\n}\n\nresource onelogin_roles test_resource {\n\tname = \"test\"\n\tapps = [1, 2, 3]\n}\n\nresource onelogin_users test_resource {\n\tusername = \"test\"\n\temail = \"test@test.test\"\n}\n\nresource aws_iam_user test_resource {\n\tpath = \"/\"\n}\n\n"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			clients := clients.Clients{
				ClientConfigs: clients.ClientConfigs{
					OneLoginClientID:     "ONELOGIN_CLIENT_ID",
					OneLoginClientSecret: "ONELOGIN_CLIENT_SECRET",
					OneLoginURL:          "ONELOGIN_OAPI_URL",
					AwsRegion:            "us-west-2",
				},
			}
			importables := tfimportables.New(&clients)
			actual := ConvertTFStateToHCL(test.InputState, importables)
			assert.Equal(t, len(test.ExpectedOutput), len(string(actual)))
		})
	}
}
