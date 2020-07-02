package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin/terraform/import"
	"github.com/onelogin/onelogin/terraform/importables"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "terraform-import",
		Short: `Import resources to local Terraform state.`,
		Long: `Uses Terraform Import to collect resources from a remote and
		create new .tfstate and .tf files so you can
		begin managing existing resources with Terraform.
		Available Imports:
			onelogin_apps          => all apps
			onelogin_saml_apps     => SAML apps only
			onelogin_oidc_apps     => OIDC apps only
			onelogin_user_mappings => user mappings`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				log.Fatalf("Must specify resource to import!")
			}
			return nil
		},
		Run: terraformImport,
	})
}

func terraformImport(cmd *cobra.Command, args []string) {
	fmt.Println("Terraform Import!")
	clientConfig := client.APIClientConfig{Timeout: 5}
	var (
		profiles      map[string]map[string]interface{}
		activeProfile map[string]interface{}
		searchId      string
	)
	if err := viper.Unmarshal(&profiles); err != nil {
		fmt.Println("No profiles detected. Add a profile with [onelogin profiles add <profile_name>]")
	} else {
		for _, prof := range profiles {
			if prof["active"].(bool) == true {
				activeProfile = prof
				break
			}
		}
		if activeProfile == nil {
			fmt.Println("No active profile detected. Authenticating with environment variables")
			clientConfig.ClientID = os.Getenv("ONELOGIN_CLIENT_ID")
			clientConfig.ClientSecret = os.Getenv("ONELOGIN_CLIENT_SECRET")
			clientConfig.Url = os.Getenv("ONELOGIN_OAPI_URL")
		} else {
			fmt.Println("Using profile", activeProfile["name"].(string))
			clientConfig.ClientID = activeProfile["client_id"].(string)
			clientConfig.ClientSecret = activeProfile["client_secret"].(string)
			clientConfig.Url = fmt.Sprintf("https://api.%s.onelogin.com", activeProfile["region"].(string))
		}
	}
	oneloginClient, err := client.NewClient(&clientConfig)
	if err != nil {
		log.Fatalln("Unable to connect to remote!", err)
	}
	if len(args) > 1 {
		searchId = args[1]
	}

	importables := map[string]tfimportables.Importable{
		"onelogin_apps":          tfimportables.OneloginAppsImportable{Service: oneloginClient.Services.AppsV2, SearchID: searchId},
		"onelogin_saml_apps":     tfimportables.OneloginAppsImportable{Service: oneloginClient.Services.AppsV2, SearchID: searchId, AppType: "onelogin_saml_apps"},
		"onelogin_oidc_apps":     tfimportables.OneloginAppsImportable{Service: oneloginClient.Services.AppsV2, SearchID: searchId, AppType: "onelogin_oidc_apps"},
		"onelogin_user_mappings": tfimportables.OneloginUserMappingsImportable{Service: oneloginClient.Services.UserMappingsV2},
	}

	importable, ok := importables[strings.ToLower(args[0])]

	if !ok {
		availableImportables := make([]string, 0, len(importables))
		for k := range importables {
			availableImportables = append(availableImportables, fmt.Sprintf("%s", k))
		}
		log.Fatalf("Available resources: %s", availableImportables)
	}

	tfimport.ImportTFStateFromRemote(importable)
}
