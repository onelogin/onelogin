package tfimportables

import (
	"testing"

	"github.com/onelogin/onelogin-go-sdk/pkg/models"
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/stretchr/testify/assert"
)

func TestAssembleResourceDefinitions(t *testing.T) {
	tests := map[string]struct {
		InputApps   []models.App
		ExpectedOut []ResourceDefinition
	}{
		"it creates a the minimum required representation of a resource in HCL": {
			InputApps: []models.App{
				models.App{
					Name:       oltypes.String("test1"),
					AuthMethod: oltypes.Int32(8),
					ID:         oltypes.Int32(1),
				},
				models.App{
					Name:       oltypes.String("test2"),
					AuthMethod: oltypes.Int32(2),
					ID:         oltypes.Int32(2),
				},
				models.App{
					Name:       oltypes.String("test3"),
					AuthMethod: oltypes.Int32(1),
					ID:         oltypes.Int32(3),
				},
			},
			ExpectedOut: []ResourceDefinition{
				ResourceDefinition{
					Provider: "onelogin",
					Type:     "onelogin_oidc_apps",
					Name:     "test1-1",
				},
				ResourceDefinition{
					Provider: "onelogin",
					Type:     "onelogin_saml_apps",
					Name:     "test2-2",
				},
				ResourceDefinition{
					Provider: "onelogin",
					Type:     "onelogin_apps",
					Name:     "test3-3",
				},
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
