package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks/envs"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockEnvVarsService struct{}

func (svc MockEnvVarsService) Query(query *smarthookenvs.SmartHookEnvVarQuery) ([]smarthookenvs.EnvVar, error) {
	return []smarthookenvs.EnvVar{
		smarthookenvs.EnvVar{Name: oltypes.String("test_1"), ID: oltypes.String("1")},
		smarthookenvs.EnvVar{Name: oltypes.String("test_2"), ID: oltypes.String("2")},
	}, nil
}

func (svc MockEnvVarsService) GetOne(id string) (*smarthookenvs.EnvVar, error) {
	return &smarthookenvs.EnvVar{Name: oltypes.String("test_1"), ID: oltypes.String("1")}, nil
}

func TestImportEnvVarFromRemote(t *testing.T) {
	tests := map[string]struct {
		SearchID   *string
		Importable OneloginSmartHookEnvVarsImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all smarthookenvs": {
			Importable: OneloginSmartHookEnvVarsImportable{Service: MockEnvVarsService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "test_1-1", ImportID: "1", Type: "onelogin_smarthook_environment_variables"},
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "test_2-2", ImportID: "2", Type: "onelogin_smarthook_environment_variables"},
			},
		},
		"It gets one smarthook": {
			SearchID:   oltypes.String("1"),
			Importable: OneloginSmartHookEnvVarsImportable{Service: MockEnvVarsService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin/onelogin", Name: "test_1-1", ImportID: "1", Type: "onelogin_smarthook_environment_variables"},
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
