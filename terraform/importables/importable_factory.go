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
		case "onelogin_users":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginUsersImportable{Service: remoteClient.Services.UsersV2}
		case "onelogin_apps", "onelogin_saml_apps", "onelogin_oidc_apps":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginAppsImportable{Service: remoteClient.Services.AppsV2, AppType: importableType}
		case "onelogin_app_rules":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginAppRulesImportable{Service: remoteClient.Services.AppRulesV2}
		case "onelogin_user_mappings":
			remoteClient := imf.Clients.OneLoginClient()
			imf.importables[importableType] = &OneloginUserMappingsImportable{Service: remoteClient.Services.UserMappingsV2}
		default:
			log.Fatalf("The importable %s is not configured", importableType)
		}
	}
	return imf.importables[importableType]
}
