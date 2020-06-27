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
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_oidc_apps", Name: "onelogin_oidc_apps-1"},
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_saml_apps", Name: "onelogin_saml_apps-2"},
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_apps", Name: "onelogin_apps-3"},
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

func TestGetAllApps(t *testing.T) {
	tests := map[string]struct {
		Importable OneloginAppsImportable
		Service    AppQuerier
		Expected   []apps.App
	}{
		"It pulls all apps of a certain type": {
			Importable: OneloginAppsImportable{AppType: "onelogin_saml_apps"},
			Service:    MockAppsService{},
			Expected: []apps.App{
				apps.App{Name: oltypes.String("test2"), AuthMethod: oltypes.Int32(2), ID: oltypes.Int32(2)},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := test.Importable.GetAllApps(test.Service)
			assert.Equal(t, test.Expected, actual)
		})
	}
}

func TestImportAppFromRemote(t *testing.T) {
	tests := map[string]struct {
		Importable OneloginAppsImportable
		Service    AppQuerier
		Expected   []ResourceDefinition
	}{
		"It pulls all apps of a certain type": {
			Importable: OneloginAppsImportable{AppType: "onelogin_saml_apps", Service: MockAppsService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test2-2", Type: "onelogin_saml_apps"},
			},
		},
		"It gets one app": {
			Importable: OneloginAppsImportable{AppType: "onelogin_saml_apps", Service: MockAppsService{}, SearchID: "1"},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test2-2", Type: "onelogin_saml_apps"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := test.Importable.ImportFromRemote()
			assert.Equal(t, test.Expected, actual)
		})
	}
}
