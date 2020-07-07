package tfimportables

import (
	"fmt"
	"log"
	"strconv"

	"github.com/onelogin/onelogin-go-sdk/pkg/services/apps"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
)

type AppQuerier interface {
	Query(query *apps.AppsQuery) ([]apps.App, error)
	GetOne(id int32) (*apps.App, error)
}

type OneloginAppsImportable struct {
	AppType  string
	SearchID string
	Service  AppQuerier
}

// Interface requirement to be an Importable. Calls out to remote (onelogin api) and
// creates their Terraform ResourceDefinitions
func (i OneloginAppsImportable) ImportFromRemote() []ResourceDefinition {
	var remoteApps []apps.App
	if i.SearchID == "" {
		fmt.Println("Collecting Apps from OneLogin...")
		remoteApps = i.GetAllApps(i.Service)
	} else {
		fmt.Printf("Collecting App %s from OneLogin...\n", i.SearchID)
		id, err := strconv.Atoi(i.SearchID)
		if err != nil {
			log.Fatalln("invalid input given for id", i.SearchID)
		}
		app, err := i.Service.GetOne(int32(id))
		if err != nil {
			log.Fatalln("Unable to locate resource with id", id)
		}
		remoteApps = []apps.App{*app}
	}
	resourceDefinitions := assembleResourceDefinitions(remoteApps)
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
		resourceDefinition.Name = fmt.Sprintf("%s-%d", utils.ToSnakeCase(resourceDefinition.Type), *app.ID)
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

func (i OneloginAppsImportable) HCLShape() interface{} {
	return &AppData{}
}

// the underlying data that represents the resource from the remote in terraform.
// add fields here so they can be unmarshalled from tfstate json into the struct and handled by the importer
type AppData struct {
	AllowAssumedSignin *bool                `json:"allow_assumed_signin,omitempty"`
	ConnectorID        *int32               `json:"connector_id,omitempty"`
	Description        *string              `json:"description,omitempty"`
	Name               *string              `json:"name,omitempty"`
	Notes              *string              `json:"notes,omitempty"`
	Visible            *bool                `json:"visible,omitempty"`
	Configuration      AppConfigurationData `json:"configuration,omitempty"`
	Provisioning       AppProvisioningData  `json:"provisioning,omitempty"`
	Parameters         []AppParametersData  `json:"parameters,omitempty"`
	Rules              []AppRuleData        `json:"rules,omitempty"`
}

// AppProvisioning is the contract for provisioning.
type AppProvisioningData struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// AppConfiguration is the contract for configuration.
type AppConfigurationData struct {
	RedirectURI                   *string `json:"redirect_uri,omitempty"`
	RefreshTokenExpirationMinutes *string `json:"refresh_token_expiration_minutes,omitempty"`
	LoginURL                      *string `json:"login_url,omitempty"`
	OidcApplicationType           *string `json:"oidc_application_type,omitempty"`
	TokenEndpointAuthMethod       *string `json:"token_endpoint_auth_method,omitempty"`
	AccessTokenExpirationMinutes  *string `json:"access_token_expiration_minutes,omitempty"`
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
