package tfimportables

import (
	"context"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssembleOktaAppResourceDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputApps   []okta.App
		ExpectedOut []ResourceDefinition
	}{
		"it creates a the minimum required representation of a resource in HCL": {
			InputApps: []okta.App{
				&okta.Application{Label: "test1", Id: "1", SignOnMode: "OPENID_CONNECT"},
				&okta.Application{Label: "test2", Id: "2", SignOnMode: "SAML_2_0"},
				&okta.Application{Label: "test3", Id: "3"},
			},
			ExpectedOut: []ResourceDefinition{
				ResourceDefinition{Provider: "oktadeveloper/okta", Type: "okta_app_oauth", ImportID: "1", Name: "okta_app_oauth_test1-1"},
				ResourceDefinition{Provider: "oktadeveloper/okta", Type: "okta_app_saml", ImportID: "2", Name: "okta_app_saml_test2-2"},
				ResourceDefinition{Provider: "oktadeveloper/okta", Type: "okta_app_basic_auth", ImportID: "3", Name: "okta_app_basic_auth_test3-3"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := assembleOktaResourceDefinitions(test.InputApps)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}

type MockOktaAppsService struct{}

func (svc MockOktaAppsService) ListApplications(context.Context, *query.Params) ([]okta.App, *okta.Response, error) {
	return []okta.App{
		&okta.Application{Label: "test2", Id: "2", SignOnMode: "SAML_2_0"},
	}, nil, nil
}

func TestImportOktaAppFromRemote(t *testing.T) {
	tests := map[string]struct {
		Importable OktaAppsImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all apps of a certain type": {
			Importable: OktaAppsImportable{Service: MockOktaAppsService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "oktadeveloper/okta", Name: "okta_app_saml_test2-2", ImportID: "2", Type: "okta_app_saml"},
			},
		},
		"It gets one app": {
			Importable: OktaAppsImportable{Service: MockOktaAppsService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "oktadeveloper/okta", Name: "okta_app_saml_test2-2", ImportID: "2", Type: "okta_app_saml"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := test.Importable.ImportFromRemote(nil)
			assert.Equal(t, test.Expected, actual)
		})
	}
}
