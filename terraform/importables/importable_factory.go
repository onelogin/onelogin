// Package tfimportables importable_factory.go
// This module creates a list of importable instances so each can be instantiated once and shared among other callers.
//
// Adding importables
// Add a case to the switch statement where the key is the name of the resource as it should be represented in terraform
// then call the requisite client method if needed. Be sure the client exists in the clients package to avoid comiler errors.
// Finally, add the importable to the map of importables so it can be fetched by referencing the terraform name via the terraform naming convention
package tfimportables

import (
	"github.com/onelogin/onelogin/clients"

	"log"
)

// ImportableList is the list of created importables referenced by a map where the key is the name used to identify it in terraform
type ImportableList struct {
	importables map[string]Importable
	Clients     *clients.Clients
}

func New(clients *clients.Clients) *ImportableList {
	imf := ImportableList{}
	imf.importables = map[string]Importable{}
	imf.Clients = clients
	return &imf
}

func (imf *ImportableList) GetImportable(importableType string) Importable {
	if imf.importables[importableType] == nil {
		switch importableType {
		case "aws_iam_user":
			remoteClient := imf.Clients.AwsIamClient()
			imf.importables[importableType] = &AWSUsersImportable{Service: remoteClient}
		case "okta_apps", "okta_app_oauth", "okta_app_saml", "okta_app_basic_auth":
			remoteClient := imf.Clients.OktaClient()
			imf.importables[importableType] = &OktaAppsImportable{Service: remoteClient.Application}
		case "onelogin_users":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginUsersImportable{Service: remoteClient.Services.UsersV2}
		case "onelogin_apps", "onelogin_saml_apps", "onelogin_oidc_apps":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginAppsImportable{Service: remoteClient.Services.AppsV2, AppType: importableType}
		case "onelogin_user_mappings":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginUserMappingsImportable{Service: remoteClient.Services.UserMappingsV2}
		case "onelogin_smarthooks":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginSmartHooksImportable{Service: remoteClient.Services.SmartHooksV1}
		case "onelogin_smarthook_env_vars", "onelogin_smarthook_environment_variables":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginSmartHookEnvVarsImportable{Service: remoteClient.Services.SmartHooksEnvVarsV1}
		case "onelogin_roles":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginRolesImportable{Service: remoteClient.Services.RolesV1}
		default:
			log.Fatalf("The importable %s is not configured", importableType)
		}
	}
	return imf.importables[importableType]
}
