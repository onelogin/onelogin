package tfimportables

import (
	"fmt"
	"log"

	"github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
)

type SmartHookQuerier interface {
	Query(query *smarthooks.SmartHookQuery) ([]smarthooks.SmartHook, error)
	GetOne(id string) (*smarthooks.SmartHook, error)
}

type OneloginSmartHooksImportable struct {
	Service SmartHookQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginSmartHooksImportable) ImportFromRemote(searchId *string) []ResourceDefinition {
	out := []smarthooks.SmartHook{}
	var err error
	if searchId == nil || *searchId == "" {
		fmt.Println("Collecting SmartHooks from OneLogin...")
		out, err = i.Service.Query(nil) // Todo, interface to pass these queries down
		if err != nil {
			log.Fatalln("Unable to get SmartHooks", err)
		}
	} else {
		fmt.Printf("Collecting SmartHook %s from OneLogin...\n", *searchId)
		smarthook, err := i.Service.GetOne(*searchId)
		if err != nil {
			log.Fatalln("Unable to locate resource with id", searchId)
		}
		out = append(out, *smarthook)
	}

	resourceDefinitions := make([]ResourceDefinition, len(out))
	for i, rd := range out {
		resourceDefinitions[i] = ResourceDefinition{
			Provider: "onelogin/onelogin",
			Type:     "onelogin_smarthooks",
			Name:     utils.ToSnakeCase(fmt.Sprintf("%s-%s", *rd.Type, utils.ReplaceSpecialChar(*rd.ID, ""))),
			ImportID: fmt.Sprintf("%s", *rd.ID),
		}
	}
	return resourceDefinitions
}

func (i OneloginSmartHooksImportable) HCLShape() interface{} {
	return &SmartHook{}
}

// SmartHook represents a OneLogin SmartHook with associated resource data
type SmartHook struct {
	ID             *string           `json:"id,omitempty"`
	Type           *string           `json:"type,omitempty"`
	Disabled       *bool             `json:"disabled,omitempty"`
	Timeout        *int32            `json:"timeout,omitempty"`
	EnvVars        []string          `json:"env_vars"`
	Runtime        *string           `json:"runtime,omitempty"`
	ContextVersion *string           `json:"context_version,omitempty"`
	Retries        *int32            `json:"retries,omitempty"`
	Options        *Options          `json:"options,omitempty"`
	Packages       map[string]string `json:"packages,omitempty"`
	Function       *string           `json:"function,omitempty"`
	Status         *string           `json:"status,omitempty"`
	Conditions     []Condition       `json:"conditions,omitempty"`
}

// SmartHookOptions represents the options to be associated with a SmartHook
type Options struct {
	RiskEnabled          *bool `json:"risk_enabled,omitempty"`
	MFADeviceInfoEnabled *bool `json:"mfa_device_info_enabled,omitempty"`
	LocationEnabled      *bool `json:"location_enabled,omitempty"`
}

type Condition struct {
	Source   *string `json:"source,omitempty"`
	Operator *string `json:"operator,omitempty"`
	Value    *string `json:"value,omitempty"`
}
