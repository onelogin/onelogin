package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssembleResourceDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputApps   []apps.App
		ExpectedOut []ResourceDefinition
	}{
		"it creates a the minimum required representation of a resource in HCL": {
			InputApps: []apps.App{
				apps.App{Name: oltypes.String("test1"), AuthMethod: oltypes.Int32(8), ID: oltypes.Int32(1)},
				apps.App{Name: oltypes.String("test2"), AuthMethod: oltypes.Int32(2), ID: oltypes.Int32(2)},
				apps.App{Name: oltypes.String("test3"), AuthMethod: oltypes.Int32(1), ID: oltypes.Int32(3)},
			},
			ExpectedOut: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin/onelogin", Type: "onelogin_oidc_apps", ImportID: "1", Name: "onelogin_oidc_apps_test1-1"},
				ResourceDefinition{Provider: "onelogin/onelogin", Type: "onelogin_saml_apps", ImportID: "2", Name: "onelogin_saml_apps_test2-2"},
				ResourceDefinition{Provider: "onelogin/onelogin", Type: "onelogin_apps", ImportID: "3", Name: "onelogin_apps_test3-3"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := assembleResourceDefinitions(test.InputApps)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}

type MockAppsService struct{}

func (svc MockAppsService) Query(query *apps.AppsQuery) ([]apps.App, error) {
	return []apps.App{
		apps.App{Name: oltypes.String("test2"), AuthMethod: oltypes.Int32(2), ID: oltypes.Int32(2)},
	}, nil
}

func (svc MockAppsService) GetOne(id int32) (*apps.App, error) {
	return &apps.App{Name: oltypes.String("test2"), AuthMethod: oltypes.Int32(2), ID: oltypes.Int32(2)}, nil
}

func TestImportAppFromRemote(t *testing.T) {
	tests := map[string]struct {
		SearchID   *string
		Importable OneloginAppsImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all apps of a certain type": {
			Importable: OneloginAppsImportable{AppType: "onelogin_saml_apps", Service: MockAppsService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "onelogin_saml_apps_test2-2", ImportID: "2", Type: "onelogin_saml_apps"},
			},
		},
		"It gets one app": {
			SearchID:   oltypes.String("2"),
			Importable: OneloginAppsImportable{AppType: "onelogin_saml_apps", Service: MockAppsService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "onelogin_saml_apps_test2-2", ImportID: "2", Type: "onelogin_saml_apps"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := test.Importable.ImportFromRemote(test.SearchID)
			assert.Equal(t, test.Expected, actual)
		})
	}
}
