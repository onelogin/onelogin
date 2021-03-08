package tfimportables

import (
	"github.com/onelogin/onelogin/clients"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetImportable(t *testing.T) {
	clientList := clients.New(clients.ClientConfigs{
		AwsRegion:            "us-test-2",
		OneLoginClientID:     "test",
		OneLoginClientSecret: "test",
		OneLoginURL:          "test.com",
		OktaOrgName:          "org",
		OktaBaseURL:          "org.org",
		OktaAPIToken:         "t0k3n",
	})
	importableNames := [7]string{
		"onelogin_apps",
		"onelogin_users",
		"onelogin_apps",
		"onelogin_user_mappings",
		"onelogin_roles",
		"okta_apps",
		"aws_iam_user",
	}
	tests := map[string]struct {
		Importables *ImportableList
	}{
		"It creates and returns importables": {
			Importables: New(clientList),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, name := range importableNames {
				importable := test.Importables.GetImportable(name)
				memoizedImportable := test.Importables.GetImportable(name)
				assert.Equal(t, test.Importables.importables[name], importable)
				assert.Equal(t, test.Importables.importables[name], memoizedImportable)
			}
		})
	}
}
