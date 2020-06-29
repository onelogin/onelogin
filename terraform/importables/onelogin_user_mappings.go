package tfimportables

import (
	"fmt"
	"log"

	"github.com/onelogin/onelogin-go-sdk/pkg/services/user_mappings"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
)

type OneloginUserMappingsImportable struct {
	Service UserMappingQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginUserMappingsImportable) ImportFromRemote() []ResourceDefinition {
	fmt.Println("Collecting User Mappings from OneLogin...")

	allApps := i.GetAll(i.Service)
	resourceDefinitions := assembleUserMappingResourceDefinitions(allApps)
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
func (i OneloginUserMappingsImportable) GetAll(userMappingsService UserMappingQuerier) []usermappings.UserMapping {
	um, err := userMappingsService.Query(&usermappings.UserMappingsQuery{})
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
