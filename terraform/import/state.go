package tfimport

import (
	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps"
)

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
	Data ResourceData `json:"attributes"`
}

// the underlying data that represents the resource from the remote in terraform.
// add fields here so they can be unmarshalled from tfstate json into the struct and handled by the importer
type ResourceData struct {
	AllowAssumedSignin *bool                   `json:"allow_assumed_signin,omitempty"`
	ConnectorID        *int32                  `json:"connector_id,omitempty"`
	Description        *string                 `json:"description,omitempty"`
	Name               *string                 `json:"name,omitempty"`
	Notes              *string                 `json:"notes,omitempty"`
	Visible            *bool                   `json:"visible,omitempty"`
	Provisioning       []apps.AppProvisioning  `json:"provisioning,omitempty"`
	Parameters         []apps.AppParameters    `json:"parameters,omitempty"`
	Configuration      []apps.AppConfiguration `json:"configuration,omitempty"`
}
