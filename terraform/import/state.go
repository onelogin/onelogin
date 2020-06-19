package tfimport

import (
	"encoding/json"
	"github.com/onelogin/onelogin/terraform/importables"
)

func sculpt(resourceType string, resourceData interface{}) interface{} {
	molds := map[string]interface{}{
		"onelogin_apps":          &tfimportables.AppData{},
		"onelogin_saml_apps":     &tfimportables.AppData{},
		"onelogin_oidc_apps":     &tfimportables.AppData{},
		"onelogin_user_mappings": &tfimportables.UserMappingData{},
	}
	b, _ := json.Marshal(resourceData)
	o := molds[resourceType]
	json.Unmarshal(b, &o)
	return o
}

// State is the in memory representation of tfstate.
type State struct {
	Resources []StateResource `json:"resources"`
}

// Terraform resource representation
type StateResource struct {
	Content   []byte
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Provider  string             `json:"provider"`
	Instances []ResourceInstance `json:"instances"`
}

// An instance of a particular resource without the terraform information
type ResourceInstance struct {
	Data interface{} `json:"attributes"`
}
