package tfimportables

import (
	"fmt"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/users"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
	"log"
	"strconv"
)

type UserQuerier interface {
	Query(query *users.UserQuery) ([]users.User, error)
	GetOne(id int32) (*users.User, error)
}

type OneloginUsersImportable struct {
	Service UserQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginUsersImportable) ImportFromRemote(searchId *string) []ResourceDefinition {
	out := []users.User{}
	var err error
	if searchId == nil || *searchId == "" {
		fmt.Println("Collecting Users from OneLogin...")
		out, err = i.Service.Query(nil) // Todo, interface to pass these queries down
		if err != nil {
			log.Fatalln("Unable to get users", err)
		}
	} else {
		fmt.Printf("Collecting User %s from OneLogin...\n", *searchId)
		id, err := strconv.Atoi(*searchId)
		if err != nil {
			log.Fatalln("invalid input given for id", *searchId)
		}
		user, err := i.Service.GetOne(int32(id))
		if err != nil {
			log.Fatalln("Unable to locate resource with id", id)
		}
		out = append(out, *user)
	}
	resourceDefinitions := make([]ResourceDefinition, len(out))
	for i, rd := range out {
		resourceDefinitions[i] = ResourceDefinition{
			Provider: "onelogin",
			Type:     "onelogin_users",
			Name:     utils.ReplaceSpecialChar(*rd.Email, "_"),
			ImportID: fmt.Sprintf("%d", *rd.ID),
		}
	}
	return resourceDefinitions
}

func (i OneloginUsersImportable) HCLShape() interface{} {
	return &UserData{}
}

// the underlying data that represents the resource from the remote in terraform.
// add fields here so they can be unmarshalled from tfstate json into the struct and handled by the importer
type UserData struct {
	Firstname            *string `json:"firstname,omitempty"`
	Lastname             *string `json:"lastname,omitempty"`
	Username             *string `json:"username,omitempty"`
	Email                *string `json:"email,omitempty"`
	DistinguishedName    *string `json:"distinguished_name,omitempty"`
	Samaccountname       *string `json:"samaccountname,omitempty"`
	UserPrincipalName    *string `json:"userprincipalname,omitempty"`
	MemberOf             *string `json:"member_of,omitempty"`
	Phone                *string `json:"phone,omitemepty"`
	Title                *string `json:"title,omitempty"`
	Company              *string `json:"company,omitempty"`
	Department           *string `json:"department,omitempty"`
	Comment              *string `json:"comment,omitempty"`
	State                *int32  `json:"state,omitempty"`
	Status               *int32  `json:"status,omitempty"`
	InvalidLoginAttempts *int32  `json:"invalid_login_attempts,omitempty"`
	GroupID              *int32  `json:"group_id,omitempty"`
	DirectoryID          *int32  `json:"directory_id,omitempty"`
	TrustedIDPID         *int32  `json:"trusted_idp_id,omitempty"`
	ManagerADID          *int32  `json:"manager_ad_id,omitempty"`
	ManagerUserID        *int32  `json:"manager_user_id,omitempty"`
	ExternalID           *int32  `json:"external_id,omitempty"`
}
