package tfimportables

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/onelogin/onelogin-go-sdk/pkg/oltypes"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockAWSUsersService struct{}

func (svc MockAWSUsersService) ListUsers(input *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	return &iam.ListUsersOutput{
		Users: []*iam.User{
			&iam.User{UserName: oltypes.String("test_1"), Path: oltypes.String("/"), UserId: oltypes.String("1")},
			&iam.User{UserName: oltypes.String("test_2"), Path: oltypes.String("/"), UserId: oltypes.String("2")},
		},
	}, nil
}

func TestImportAWSUserFromRemote(t *testing.T) {
	tests := map[string]struct {
		Importable AWSUsersImportable
		Expected   []ResourceDefinition
	}{
		"It pulls all users": {
			Importable: AWSUsersImportable{Service: MockAWSUsersService{}},
			Expected: []ResourceDefinition{
				ResourceDefinition{Provider: "aws", Name: "test_1", ImportID: "test_1", Type: "aws_iam_user"},
				ResourceDefinition{Provider: "aws", Name: "test_2", ImportID: "test_2", Type: "aws_iam_user"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := test.Importable.ImportFromRemote(nil)
			assert.Equal(t, test.Expected, actual)
		})
	}
}
