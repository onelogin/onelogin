package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceSpecialChar(t *testing.T) {
	tests := map[string]struct {
		InputStr    string
		InputRep    string
		ExpectedOut string
	}{
		"It replaces the non alpha-numeric with the specified character": {
			InputStr:    "stuff+test",
			InputRep:    "&",
			ExpectedOut: "stuff&test",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := ReplaceSpecialChar(test.InputStr, test.InputRep)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := map[string]struct {
		InputStr    string
		ExpectedOut string
	}{
		"It converts the PascalCase to snake_case": {
			InputStr:    "PascalCase",
			ExpectedOut: "pascal_case",
		},
		"It converts the cascalCase to snake_case": {
			InputStr:    "camelCase",
			ExpectedOut: "camel_case",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := ToSnakeCase(test.InputStr)
			assert.Equal(t, test.ExpectedOut, actual)
		})
	}
}
