package tfimportables

import (
	"context"
	"fmt"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
	"time"
)

type OktaAppQuerier interface {
	ListApplications(context.Context, *query.Params) ([]okta.App, *okta.Response, error)
}

type OktaAppsImportable struct {
	SearchID string
	Service  OktaAppQuerier
}

func (i OktaAppsImportable) ImportFromRemote() []ResourceDefinition {
	apps := i.getAllApps()
	rd := assembleOktaResourceDefinitions(apps)
	return rd
}

func assembleOktaResourceDefinitions(allApps []okta.App) []ResourceDefinition {
	resourceDefinitions := make([]ResourceDefinition, len(allApps))
	for i, a := range allApps {
		resourceDefinition := ResourceDefinition{Provider: "okta"}
		// there's more but this is good 'nuff for the hackathon lol
		switch a.(*okta.Application).SignOnMode {
		case "OPENID_CONNECT":
			resourceDefinition.Type = "okta_app_oauth"
		case "SAML_2_0":
			resourceDefinition.Type = "okta_app_saml"
		default:
			resourceDefinition.Type = "okta_app_basic_auth"
		}
		resourceDefinition.Name = fmt.Sprintf("%s-%s", utils.ToSnakeCase(resourceDefinition.Type), a.(*okta.Application).Id)
		resourceDefinitions[i] = resourceDefinition
	}
	return resourceDefinitions
}

func (i OktaAppsImportable) getAllApps() []okta.App {
	apps, resp, err := i.Service.ListApplications(context.TODO(), nil)
	if err != nil {
		fmt.Println("ERROR", err)
		fmt.Println("RESP", resp)
	}
	for _, a := range apps {
		fmt.Println(a.(*okta.Application).SignOnMode)
	}
	return apps
}

func (i OktaAppsImportable) HCLShape() interface{} {
	return &OktaAppData{}
}

// TODO what fields do we need to snag for saml / oidc ?
type OktaAppData struct {
	Embedded      interface{}                    `json:"_embedded,omitempty"`
	Links         interface{}                    `json:"_links,omitempty"`
	Accessibility *okta.ApplicationAccessibility `json:"accessibility,omitempty"`
	Created       *time.Time                     `json:"created,omitempty"`
	Credentials   *okta.ApplicationCredentials   `json:"credentials,omitempty"`
	Features      []string                       `json:"features,omitempty"`
	Id            string                         `json:"id,omitempty"`
	Label         string                         `json:"label,omitempty"`
	LastUpdated   *time.Time                     `json:"lastUpdated,omitempty"`
	Licensing     *okta.ApplicationLicensing     `json:"licensing,omitempty"`
	Name          string                         `json:"name,omitempty"`
	Profile       interface{}                    `json:"profile,omitempty"`
	Settings      *okta.ApplicationSettings      `json:"settings,omitempty"`
	SignOnMode    string                         `json:"signOnMode,omitempty"`
	Status        string                         `json:"status,omitempty"`
	Visibility    *okta.ApplicationVisibility    `json:"visibility,omitempty"`
}
