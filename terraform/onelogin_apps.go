package terraform

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/onelogin/onelogin-cli/utils"
	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin-go-sdk/pkg/models"
)

type OneloginAppsImportable struct {
	AppType string
}

func (i OneloginAppsImportable) ImportResourceDefinitionsFromRemote() []ResourceDefinition {
	fmt.Println("Collecting Apps from OneLogin...")

	allApps := getAllApps(i.AppType)

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
		resourceDefinition.PrepareTFImport()
		resourceDefinitions[i] = resourceDefinition
	}
	return resourceDefinitions
}

func getAllApps(appType string) []models.App {
	var (
		resp    *http.Response
		apps    []models.App
		allApps []models.App
		err     error
		next    string
	)

	appTypeQueryMap := map[string]string{
		"onelogin_apps":      "",
		"onelogin_saml_apps": "2",
		"onelogin_oidc_apps": "8",
	}
	requestedAppType := appTypeQueryMap[appType]

	sdkClient, _ := client.NewClient(&client.APIClientConfig{
		Timeout:      5,
		ClientID:     os.Getenv("ONELOGIN_CLIENT_ID"),
		ClientSecret: os.Getenv("ONELOGIN_CLIENT_SECRET"),
		Url:          os.Getenv("ONELOGIN_OAPI_URL"),
	})

	resp, apps, err = sdkClient.Services.AppsV2.GetApps(&models.AppsQuery{
		AuthMethod: requestedAppType,
	})

	for {
		allApps = append(allApps, apps...)
		next = resp.Header.Get("After-Cursor")
		if next == "" || err != nil {
			break
		}
		resp, apps, err = sdkClient.Services.AppsV2.GetApps(&models.AppsQuery{
			AuthMethod: requestedAppType,
			Cursor:     next,
		})
	}
	if err != nil {
		log.Fatal("error retrieving apps ", err)
	}
	return allApps
}
