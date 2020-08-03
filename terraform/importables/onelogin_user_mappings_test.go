package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/user_mappings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssembleUserMappingResourceDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputUserMappings []usermappings.UserMapping
		ExpectedOut       []ResourceDefinition
	}{
		"it creates a the minimum required representation of a resource in HCL": {
			InputUserMappings: []usermappings.UserMapping{
				usermappings.UserMapping{Name: oltypes.String("test1"), ID: oltypes.Int32(1)},
				usermappings.UserMapping{Name: oltypes.String("test2"), ID: oltypes.Int32(2)},
				usermappings.UserMapping{Name: oltypes.String("test3"), ID: oltypes.Int32(3)},
			},
			ExpectedOut: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_user_mappings", ImportID: "1", Name: "test1"},
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_user_mappings", ImportID: "2", Name: "test2"},
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_user_mappings", ImportID: "3", Name: "test3"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := assembleUserMappingResourceDefinitions(test.InputUserMappings)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}

type MockUserMappingService struct{}

func (svc MockUserMappingService) Query(query *usermappings.UserMappingsQuery) ([]usermappings.UserMapping, error) {
	return []usermappings.UserMapping{
		usermappings.UserMapping{Name: oltypes.String("test2"), ID: oltypes.Int32(2)},
	}, nil
}

func (svc MockUserMappingService) GetOne(id int32) (*usermappings.UserMapping, error) {
	return &usermappings.UserMapping{Name: oltypes.String("test2"), ID: oltypes.Int32(2)}, nil
}

func TestImportUserMappingFromRemote(t *testing.T) {
	tests := map[string]struct {
		Importable OneloginUserMappingsImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all user mappings of a certain type": {
			Importable: OneloginUserMappingsImportable{Service: MockUserMappingService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test2", ImportID: "2", Type: "onelogin_user_mappings"},
			},
		},
		"It gets one user mapping": {
			Importable: OneloginUserMappingsImportable{Service: MockUserMappingService{}, SearchID: "1"},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test2", ImportID: "2", Type: "onelogin_user_mappings"},
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
