package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/roles"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockRolesService struct{}

func (svc MockRolesService) Query(query *roles.RoleQuery) ([]roles.Role, error) {
	return []roles.Role{
		roles.Role{Name: oltypes.String("test_1"), Apps: []int32{1, 2, 3}, ID: oltypes.Int32(1)},
		roles.Role{Name: oltypes.String("test_2"), Apps: []int32{1, 2, 3}, ID: oltypes.Int32(2)},
	}, nil
}

func (svc MockRolesService) GetOne(id int32) (*roles.Role, error) {
	return &roles.Role{Name: oltypes.String("test"), Apps: []int32{1, 2, 3}, ID: oltypes.Int32(1)}, nil
}

func TestImportRoleFromRemote(t *testing.T) {
	tests := map[string]struct {
		SearchID   *string
		Importable OneloginRolesImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all roles": {
			Importable: OneloginRolesImportable{Service: MockRolesService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test1", ImportID: "1", Type: "onelogin_roles"},
				ResourceDefinition{Provider: "onelogin", Name: "test2", ImportID: "2", Type: "onelogin_roles"},
			},
		},
		"It gets one role": {
			SearchID:   oltypes.String("1"),
			Importable: OneloginRolesImportable{Service: MockRolesService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test", ImportID: "1", Type: "onelogin_roles"},
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
