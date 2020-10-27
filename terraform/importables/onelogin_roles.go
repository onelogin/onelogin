package tfimportables

import (
	"fmt"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/roles"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
	"log"
	"strconv"
)

type RoleQuerier interface {
	Query(query *roles.RoleQuery) ([]roles.Role, error)
	GetOne(id int32) (*roles.Role, error)
}

type OneloginRolesImportable struct {
	Service RoleQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginRolesImportable) ImportFromRemote(searchId *string) []ResourceDefinition {
	out := []roles.Role{}
	var err error
	if searchId == nil || *searchId == "" {
		fmt.Println("Collecting Roles from OneLogin...")
		out, err = i.Service.Query(nil) // Todo, interface to pass these queries down
		if err != nil {
			log.Fatalln("Unable to get roles", err)
		}
	} else {
		fmt.Printf("Collecting Role %s from OneLogin...\n", *searchId)
		id, err := strconv.Atoi(*searchId)
		if err != nil {
			log.Fatalln("invalid input given for id", *searchId)
		}
		role, err := i.Service.GetOne(int32(id))
		if err != nil {
			log.Fatalln("Unable to locate resource with id", id)
		}
		out = append(out, *role)
	}
	resourceDefinitions := make([]ResourceDefinition, len(out))
	for i, rd := range out {
		resourceDefinitions[i] = ResourceDefinition{
			Provider: "onelogin",
			Type:     "onelogin_roles",
			Name:     utils.ToSnakeCase(utils.ReplaceSpecialChar(*rd.Name, "")),
			ImportID: fmt.Sprintf("%d", *rd.ID),
		}
	}
	return resourceDefinitions
}

func (i OneloginRolesImportable) HCLShape() interface{} {
	return &roles.Role{}
}
