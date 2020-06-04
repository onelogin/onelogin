package tfimportables

import (
	"fmt"
	"log"

	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
)

type OneloginAppsImportable struct {
	AppType string
	Service AppQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginAppsImportable) ImportFromRemote() []ResourceDefinition {
	fmt.Println("Collecting Apps from OneLogin...")

	allApps := i.GetAllApps(i.Service)
	resourceDefinitions := assembleResourceDefinitions(allApps)
	return resourceDefinitions
}

// helper for packing apps into ResourceDefinitions
func assembleResourceDefinitions(allApps []apps.App) []ResourceDefinition {
	resourceDefinitions := make([]ResourceDefinition, len(allApps))
	for i, app := range allApps {
		resourceDefinition := ResourceDefinition{Provider: "onelogin"}
		switch *app.AuthMethod {
		case 8:
			resourceDefinition.Type = "onelogin_oidc_apps"
		case 2:
			resourceDefinition.Type = "onelogin_saml_apps"
		default:
			resourceDefinition.Type = "onelogin_apps"
		}
		resourceDefinition.Name = fmt.Sprintf("%s-%d", utils.ToSnakeCase(utils.ReplaceSpecialChar(*app.Name, "")), *app.ID)
		resourceDefinitions[i] = resourceDefinition
	}
	return resourceDefinitions
}

// Makes the HTTP call to the remote to get the apps using the given query parameters
func (i OneloginAppsImportable) GetAllApps(appsService AppQuerier) []apps.App {

	appTypeQueryMap := map[string]string{
		"onelogin_apps":      "",
		"onelogin_saml_apps": "2",
		"onelogin_oidc_apps": "8",
	}
	requestedAppType := appTypeQueryMap[i.AppType]

	appApps, err := appsService.Query(&apps.AppsQuery{
		AuthMethod: requestedAppType,
	})
	if err != nil {
		log.Fatal("error retrieving apps ", err)
	}

	return appApps
}
