package tfimportables

import (
	"fmt"
	"github.com/onelogin/onelogin-go-sdk/pkg/services/user_mappings"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
	"log"
	"strconv"
)

type UserMappingQuerier interface {
	Query(query *usermappings.UserMappingsQuery) ([]usermappings.UserMapping, error)
	GetOne(id int32) (*usermappings.UserMapping, error)
}

type OneloginUserMappingsImportable struct {
	Service  UserMappingQuerier
	SearchID string
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginUserMappingsImportable) ImportFromRemote() []ResourceDefinition {
	var remoteUserMappings []usermappings.UserMapping
	if i.SearchID == "" {
		fmt.Println("Collecting User Mappings from OneLogin...")
		remoteUserMappings = i.getOneLoginUserMappings()
	} else {
		fmt.Printf("Collecting User Mapping %s from OneLogin...\n", i.SearchID)
		id, err := strconv.Atoi(i.SearchID)
		if err != nil {
			log.Fatalln("invalid input given for id", i.SearchID)
		}
		userMapping, err := i.Service.GetOne(int32(id))
		if err != nil {
			log.Fatalln("Unable to locate resource with id", id)
		}
		remoteUserMappings = []usermappings.UserMapping{*userMapping}
	}
	resourceDefinitions := assembleUserMappingResourceDefinitions(remoteUserMappings)
	return resourceDefinitions
}

// helper for packing apps into ResourceDefinitions
func assembleUserMappingResourceDefinitions(allUserMappings []usermappings.UserMapping) []ResourceDefinition {
	resourceDefinitions := make([]ResourceDefinition, len(allUserMappings))
	for i, userMapping := range allUserMappings {
		resourceDefinition := ResourceDefinition{Provider: "onelogin"}
		resourceDefinition.Type = "onelogin_user_mappings"
		resourceDefinition.Name = fmt.Sprintf("%s-%d", utils.ToSnakeCase(utils.ReplaceSpecialChar(*userMapping.Name, "")), *userMapping.ID)
		resourceDefinitions[i] = resourceDefinition
	}
	return resourceDefinitions
}

// Makes the HTTP call to the remote to get the apps using the given query parameters
func (i OneloginUserMappingsImportable) getOneLoginUserMappings() []usermappings.UserMapping {
	um, err := i.Service.Query(&usermappings.UserMappingsQuery{})
	if err != nil {
		log.Fatal("error retrieving apps ", err)
	}
	return um
}

func (i OneloginUserMappingsImportable) HCLShape() interface{} {
	return &UserMappingData{}
}

// the underlying data that represents the resource from the remote in terraform.
// add fields here so they can be unmarshalled from tfstate json into the struct and handled by the importer
type UserMappingData struct {
	Name       *string                     `json:"name,omitempty"`
	Match      *string                     `json:"match,omitempty"`
	Position   *int32                      `json:"position,omitempty"`
	Enabled    *bool                       `json:"enabled,omitempty"`
	Conditions []UserMappingConditionsData `json:"conditions,omitempty"` // we managed to get lucky thus far but if multiple resources have the same field and theyre different types this will be a problem
	Actions    []UserMappingActionsData    `json:"actions,omitempty"`
}

// UserMappingConditions is the contract for User Mapping Conditions.
type UserMappingConditionsData struct {
	Source   *string `json:"source,omitempty"`
	Operator *string `json:"operator,omitempty"`
	Value    *string `json:"value,omitempty"`
}

// UserMappingActions is the contract for User Mapping Actions.
type UserMappingActionsData struct {
	Action *string  `json:"action,omitempty"`
	Value  []string `json:"value,omitempty"`
}
