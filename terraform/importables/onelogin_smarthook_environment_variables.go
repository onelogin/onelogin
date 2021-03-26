package tfimportables

import (
	"fmt"
	"log"

	"github.com/onelogin/onelogin-go-sdk/pkg/services/smarthooks/envs"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
)

type SmartHookEnvVarQuerier interface {
	Query(query *smarthookenvs.SmartHookEnvVarQuery) ([]smarthookenvs.EnvVar, error)
	GetOne(id string) (*smarthookenvs.EnvVar, error)
}

type OneloginSmartHookEnvVarsImportable struct {
	Service SmartHookEnvVarQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginSmartHookEnvVarsImportable) ImportFromRemote(searchId *string) []ResourceDefinition {
	out := []smarthookenvs.EnvVar{}
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
			Type:     "onelogin_smarthook_environment_variables",
			Name:     utils.ToSnakeCase(fmt.Sprintf("%s-%s", *rd.Name, utils.ReplaceSpecialChar(*rd.ID, ""))),
			ImportID: fmt.Sprintf("%s", *rd.ID),
		}
	}
	return resourceDefinitions
}

func (i OneloginSmartHookEnvVarsImportable) HCLShape() interface{} {
	return &smarthookenvs.EnvVar{}
}
