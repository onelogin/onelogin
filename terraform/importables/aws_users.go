package tfimportables

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"log"
)

type AWSUserQuerier interface {
	ListUsers(input *iam.ListUsersInput) (*iam.ListUsersOutput, error)
}

type AWSUsersImportable struct {
	Service AWSUserQuerier
}

// Interface requirement to be an Importable. Calls out to remote (aws api) and
// creates their Terraform ResourceDefinitions
func (i AWSUsersImportable) ImportFromRemote() []ResourceDefinition {
	usrs, err := i.Service.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		log.Fatalln("There was a problem getting users", err)
	}
	out := make([]ResourceDefinition, len(usrs.Users))
	for i, u := range usrs.Users {
		out[i] = ResourceDefinition{
			Provider: "aws",
			Type:     "aws_iam_user",
			Name:     *u.UserName,
		}
	}
	return out
}

func (i AWSUsersImportable) HCLShape(outHCLShapeOption string) interface{} {
	return &AWSUserData{}
}

// the underlying data that represents the resource from the remote in terraform.
// add fields here so they can be unmarshalled from tfstate json into the struct and handled by the importer
type AWSUserData struct {
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}

type AWSOneLoginUserData struct {
	Username string `json:"name,omitempty"`
}
