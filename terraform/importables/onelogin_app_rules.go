package tfimportables

import (
	"fmt"
	"log"
	"strconv"

	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps/app_rules"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
)

type AppRuleQuerier interface {
	Query(query *apprules.AppRuleQuery) ([]apprules.AppRule, error)
	GetOne(appId int32, id int32) (*apprules.AppRule, error)
}

type OneloginAppRulesImportable struct {
	Service AppRuleQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginAppRulesImportable) ImportFromRemote(searchId *string) []ResourceDefinition {
	var remoteAppRules []apprules.AppRule
	if searchId == nil || *searchId == "" {
		fmt.Println("Collecting App Rules from OneLogin...")
		remoteAppRules = i.getOneLoginAppRules()
	} else {
		fmt.Printf("Collecting App Rule %s from OneLogin...\n", *searchId)
		id, err := strconv.Atoi(*searchId)
		if err != nil {
			log.Fatalln("invalid input given for id", *searchId)
		}
		appRule, err := i.Service.GetOne(int32(id))
		if err != nil {
			log.Fatalln("Unable to locate resource with id", id)
		}
		remoteAppRules = []apprules.AppRule{*appRule}
	}
	resourceDefinitions := assembleAppRuleResourceDefinitions(remoteAppRules)
	return resourceDefinitions
}

// Makes the HTTP call to the remote to get the app rules using the given query parameters
func (i OneloginAppRulesImportable) getOneLoginAppRules() []apprules.AppRule {
	ar, err := i.Service.Query(&apprules.AppRuleQuery{})
	if err != nil {
		log.Fatal("error retrieving app rules ", err)
	}
	return ar
}

func (i OneloginAppRulesImportable) HCLShape() interface{} {
	return &AppRuleData{}
}

// helper for packing apps into ResourceDefinitions
func assembleAppRuleResourceDefinitions(allAppRules []apprules.AppRule) []ResourceDefinition {
	resourceDefinitions := make([]ResourceDefinition, len(allAppRules))
	for i, appRule := range allAppRules {
		resourceDefinitions[i] = ResourceDefinition{
			Provider: "onelogin",
			Type:     "onelogin_app_rules",
			ImportID: fmt.Sprintf("%d", *appRule.ID),
			Name:     utils.ReplaceSpecialChar(*appRule.Name, ""),
		}
	}
	return resourceDefinitions
}

// the underlying data that represents the resource from the remote in terraform.
// add fields here so they can be unmarshalled from tfstate json into the struct and handled by the importer
type AppRuleData struct {
	Name       *string                 `json:"name,omitempty"`
	Match      *string                 `json:"match,omitempty"`
	Position   *int32                  `json:"position,omitempty"`
	Enabled    *bool                   `json:"enabled,omitempty"`
	Conditions []AppRuleConditionsData `json:"conditions,omitempty"` // we managed to get lucky thus far but if multiple resources have the same field and theyre different types this will be a problem
	Actions    []AppRuleActionsData    `json:"actions,omitempty"`
}

// AppRuleConditions is the contract for App Rule Conditions.
type AppRuleConditionsData struct {
	Source   *string `json:"source,omitempty"`
	Operator *string `json:"operator,omitempty"`
	Value    *string `json:"value,omitempty"`
}

// AppRuleActions is the contract for App Rule Actions.
type AppRuleActionsData struct {
	Action     *string  `json:"action,omitempty"`
	Value      []string `json:"value,omitempty"`
	Expression *string  `json:"expression,omitempty"`
}
