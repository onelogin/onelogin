package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/users"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockUsersService struct{}

func (svc MockUsersService) Query(query *users.UserQuery) ([]users.User, error) {
	return []users.User{
		users.User{Username: oltypes.String("test_1"), Email: oltypes.String("test_1@test.com"), ID: oltypes.Int32(1)},
		users.User{Username: oltypes.String("test_2"), Email: oltypes.String("test_2@test.com"), ID: oltypes.Int32(2)},
	}, nil
}

func (svc MockUsersService) GetOne(id int32) (*users.User, error) {
	return &users.User{Username: oltypes.String("test"), Email: oltypes.String("test@test.com"), ID: oltypes.Int32(1)}, nil
}

func TestImportUserFromRemote(t *testing.T) {
	tests := map[string]struct {
		Importable OneloginUsersImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all users": {
			Importable: OneloginUsersImportable{Service: MockUsersService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test_1_test_com", ImportID: "1", Type: "onelogin_users"},
				ResourceDefinition{Provider: "onelogin", Name: "test_2_test_com", ImportID: "2", Type: "onelogin_users"},
			},
		},
		"It gets one user": {
			Importable: OneloginUsersImportable{Service: MockUsersService{}, SearchID: "1"},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test_test_com", ImportID: "1", Type: "onelogin_users"},
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
