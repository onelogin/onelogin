package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/onelogin/onelogin-cli/terraform"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "terraform-import",
		Short: `Import resources to local Terraform state.`,
		Long: `Uses Terraform Import to collect resources from a remote and
		create new .tfstate and .tf files so you can
		begin managing existing resources with Terraform.
		Available Imports:
			onelogin_apps      => all apps
			onelogin_saml_apps => SAML apps only
			onelogin_oidc_apps => OIDC apps only`,
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

	importables := map[string]terraform.Importable{
		"onelogin_apps":      terraform.OneloginAppsImportable{},
		"onelogin_saml_apps": terraform.OneloginAppsImportable{AppType: "onelogin_saml_apps"},
		"onelogin_oidc_apps": terraform.OneloginAppsImportable{AppType: "onelogin_oidc_apps"},
	}

	importable, ok := importables[strings.ToLower(args[0])]

	if !ok {
		availableImportables := make([]string, 0, len(importables))
		for k := range importables {
			availableImportables = append(availableImportables, fmt.Sprintf("%s", k))
		}
		log.Println("Must specify resource to import!")
		log.Fatalf("Available resources: %s", availableImportables)
	}

	terraform.ImportTFStateFromRemote(importable)
	terraform.UpdateMainTFFromState()
}
