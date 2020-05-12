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

func CreateImportShells() []Shell {
	fmt.Println("Collecting Apps from OneLogin...")

	allApps := getAllApps()

	shells := make([]Shell, len(allApps))

	for i, app := range allApps {
		shell := Shell{Provider: "onelogin"}
		switch *app.AuthMethod {
		case 8:
			shell.Type = "onelogin_oidc_apps"
		case 2:
			shell.Type = "onelogin_saml_apps"
		default:
			shell.Type = "onelogin_apps"
		}
		shell.Name = fmt.Sprintf("%s-%d", utils.ToSnakeCase(utils.ReplaceSpecialChar(*app.Name, "")), *app.ID)
		shell.PrepareTFImport()
		shells[i] = shell
	}
	return shells
}

func getAllApps() []models.App {
	var (
		resp    *http.Response
		apps    []models.App
		allApps []models.App
		err     error
		next    string
	)

	sdkClient, _ := client.NewClient(&client.APIClientConfig{
		Timeout:      5,
		ClientID:     os.Getenv("ONELOGIN_CLIENT_ID"),
		ClientSecret: os.Getenv("ONELOGIN_CLIENT_SECRET"),
		Url:          os.Getenv("ONELOGIN_OAPI_URL"),
	})

	resp, apps, err = sdkClient.Services.AppsV2.GetApps(&models.AppsQuery{})

	for {
		allApps = append(allApps, apps...)
		next = resp.Header.Get("After-Cursor")
		if next == "" || err != nil {
			break
		}
		resp, apps, err = sdkClient.Services.AppsV2.GetApps(&models.AppsQuery{
			Cursor: next,
		})
	}
	if err != nil {
		log.Fatal("error retrieving apps ", err)
	}
	return allApps
}
