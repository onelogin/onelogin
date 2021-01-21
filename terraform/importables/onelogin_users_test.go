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
		SearchID   *string
		Importable OneloginUsersImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all apps of a certain type": {
			Importable: OneloginUsersImportable{Service: MockUsersService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "test_1_test", ImportID: "1", Type: "onelogin_users"},
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "test_2_test", ImportID: "2", Type: "onelogin_users"},
			},
		},
		"It gets one app": {
			SearchID:   oltypes.String("1"),
			Importable: OneloginUsersImportable{Service: MockUsersService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "test_test", ImportID: "1", Type: "onelogin_users"},
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
