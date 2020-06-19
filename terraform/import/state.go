package tfimport

import (
	"encoding/json"
)

func sculpt(resourceType string, resourceData interface{}) interface{} {
	molds := map[string]interface{}{
		"onelogin_apps":          &AppData{},
		"onelogin_saml_apps":     &AppData{},
		"onelogin_oidc_apps":     &AppData{},
		"onelogin_user_mappings": &UserMappingData{},
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

// the underlying data that represents the resource from the remote in terraform.
// add fields here so they can be unmarshalled from tfstate json into the struct and handled by the importer
type AppData struct {
	AllowAssumedSignin *bool                  `json:"allow_assumed_signin,omitempty"`
	ConnectorID        *int32                 `json:"connector_id,omitempty"`
	Description        *string                `json:"description,omitempty"`
	Name               *string                `json:"name,omitempty"`
	Notes              *string                `json:"notes,omitempty"`
	Visible            *bool                  `json:"visible,omitempty"`
	Provisioning       []AppProvisioningData  `json:"provisioning,omitempty"`
	Parameters         []AppParametersData    `json:"parameters,omitempty"`
	Configuration      []AppConfigurationData `json:"configuration,omitempty"`
	Rules              []AppRuleData          `json:"rules,omitempty"`
}

// AppProvisioning is the contract for provisioning.
type AppProvisioningData struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// AppConfiguration is the contract for configuration.
type AppConfigurationData struct {
	RedirectURI                   *string `json:"redirect_uri,omitempty"`
	RefreshTokenExpirationMinutes *int32  `json:"refresh_token_expiration_minutes,omitempty"`
	LoginURL                      *string `json:"login_url,omitempty"`
	OidcApplicationType           *int32  `json:"oidc_application_type,omitempty"`
	TokenEndpointAuthMethod       *int32  `json:"token_endpoint_auth_method,omitempty"`
	AccessTokenExpirationMinutes  *int32  `json:"access_token_expiration_minutes,omitempty"`
	ProviderArn                   *string `json:"provider_arn,omitempty"`
	SignatureAlgorithm            *string `json:"signature_algorithm,omitempty"`
}

// AppParameters is the contract for parameters.
type AppParametersData struct {
	ID                        *int32  `json:"id,omitempty"`
	Label                     *string `json:"label,omitempty"`
	UserAttributeMappings     *string `json:"user_attribute_mappings,omitempty"`
	UserAttributeMacros       *string `json:"user_attribute_macros,omitempty"`
	AttributesTransformations *string `json:"attributes_transformations,omitempty"`
	SkipIfBlank               *bool   `json:"skip_if_blank,omitempty"`
	Values                    *string `json:"values,omitempty,omitempty"`
	DefaultValues             *string `json:"default_values,omitempty"`
	ParamKeyName              *string `json:"param_key_name,omitempty"`
	ProvisionedEntitlements   *bool   `json:"provisioned_entitlements,omitempty"`
	SafeEntitlementsEnabled   *bool   `json:"safe_entitlements_enabled,omitempty"`
	IncludeInSamlAssertion    *bool   `json:"include_in_saml_assertion,omitempty"`
}

// Define our own version of the app rules to refine what fields get written to main.tf plan
type AppRuleData struct {
	Name       *string                 `json:"name,omitempty"`
	Match      *string                 `json:"match,omitempty"`
	Enabled    *bool                   `json:"enabled,omitempty"`
	Conditions []AppRuleConditionsData `json:"conditions,omitempty"`
	Actions    []AppRuleActionsData    `json:"actions,omitempty"`
}

type AppRuleActionsData struct {
	Action     *string  `json:"action,omitempty"`
	Value      []string `json:"value,omitempty"`
	Expression *string  `json:"expression,omitempty"`
}

type AppRuleConditionsData struct {
	Source   *string `json:"source,omitempty"`
	Operator *string `json:"operator,omitempty"`
	Value    *string `json:"value,omitempty"`
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
