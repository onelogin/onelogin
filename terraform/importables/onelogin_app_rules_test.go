package tfimportables

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps/app_rules"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssembleAppRuleResourceDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputAppRules []apprules.AppRule
		ExpectedOut   []ResourceDefinition
	}{
		"it creates a the minimum required representation of a resource in HCL": {
			InputAppRules: []apprules.AppRule{
				apprules.AppRule{Name: oltypes.String("test1"), ID: oltypes.Int32(1)},
				apprules.AppRule{Name: oltypes.String("test2"), ID: oltypes.Int32(2)},
				apprules.AppRule{Name: oltypes.String("test3"), ID: oltypes.Int32(3)},
			},
			ExpectedOut: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_app_rules", ImportID: "1", Name: "test1"},
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_app_rules", ImportID: "2", Name: "test2"},
				ResourceDefinition{Provider: "onelogin", Type: "onelogin_app_rules", ImportID: "3", Name: "test3"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := assembleAppRuleResourceDefinitions(test.InputAppRules)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}

type MockAppRuleService struct{}

func (svc MockAppRuleService) Query(query *apprules.AppRuleQuery) ([]apprules.AppRule, error) {
	return []apprules.AppRule{
		apprules.AppRule{Name: oltypes.String("test2"), ID: oltypes.Int32(2)},
	}, nil
}

func (svc MockAppRuleService) GetOne(id int32) (*apprules.AppRule, error) {
	return &apprules.AppRule{Name: oltypes.String("test2"), ID: oltypes.Int32(2)}, nil
}

func TestImportAppRuleFromRemote(t *testing.T) {
	tests := map[string]struct {
		SearchID   *string
		Importable OneloginAppRulesImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all apps of a certain type": {
			Importable: OneloginAppRulesImportable{Service: MockAppRuleService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test2", ImportID: "2", Type: "onelogin_app_rules"},
			},
		},
		"It gets one app": {
			SearchID:   oltypes.String("1"),
			Importable: OneloginAppRulesImportable{Service: MockAppRuleService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "onelogin", Name: "test2", ImportID: "2", Type: "onelogin_app_rules"},
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
