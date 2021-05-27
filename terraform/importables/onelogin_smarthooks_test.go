package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockSmartHooksService struct{}

func (svc MockSmartHooksService) Query(query *smarthooks.SmartHookQuery) ([]smarthooks.SmartHook, error) {
	return []smarthooks.SmartHook{
		{Function: oltypes.String("test_1"), Type: oltypes.String("test"), ID: oltypes.String("1")},
		{Function: oltypes.String("test_2"), Type: oltypes.String("test"), ID: oltypes.String("2")},
	}, nil
}

func (svc MockSmartHooksService) GetOne(id string) (*smarthooks.SmartHook, error) {
	return &smarthooks.SmartHook{Function: oltypes.String("test_1"), Type: oltypes.String("test"), ID: oltypes.String("1")}, nil
}

func TestImportSmartHookFromRemote(t *testing.T) {
	tests := map[string]struct {
		SearchID   *string
		Importable OneloginSmartHooksImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all smarthooks": {
			Importable: OneloginSmartHooksImportable{Service: MockSmartHooksService{}},
			Expected: []ResourceDefinition{
				{Provider: "onelogin/onelogin", Name: "test-1", ImportID: "1", Type: "onelogin_smarthooks"},
				{Provider: "onelogin/onelogin", Name: "test-2", ImportID: "2", Type: "onelogin_smarthooks"},
			},
		},
		"It gets one smarthook": {
			SearchID:   oltypes.String("1"),
			Importable: OneloginSmartHooksImportable{Service: MockSmartHooksService{}},
			Expected: []ResourceDefinition{
				{Provider: "onelogin/onelogin", Name: "test-1", ImportID: "1", Type: "onelogin_smarthooks"},
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
