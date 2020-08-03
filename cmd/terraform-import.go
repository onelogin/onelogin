package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin/profiles"
	"github.com/onelogin/onelogin/terraform/import"
	"github.com/onelogin/onelogin/terraform/importables"
	"github.com/onelogin/onelogin/terraform/state_parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	var (
		autoApprove    *bool
		searchId       *string
		profileService profiles.ProfileService
	)
	var tfImportCommand = &cobra.Command{
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
		Args: cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			configFile, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR, 0600)
			if err != nil {
				configFile.Close()
				log.Fatalln("Unable to open profiles file", err)
			}
			profileService = profiles.ProfileService{Repository: profiles.FileRepository{StorageMedia: configFile}}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			clientConfig := client.APIClientConfig{Timeout: 5}
			profile := profileService.GetActive()
			if profile == nil {
				fmt.Println("No active profile detected. Authenticating with environment variables")
				clientConfig.ClientID = os.Getenv("ONELOGIN_CLIENT_ID")
				clientConfig.ClientSecret = os.Getenv("ONELOGIN_CLIENT_SECRET")
				clientConfig.Url = os.Getenv("ONELOGIN_OAPI_URL")
			} else {
				fmt.Println("Using profile", (*profile).Name)
				clientConfig.ClientID = (*profile).ClientID
				clientConfig.ClientSecret = (*profile).ClientSecret
				clientConfig.Url = fmt.Sprintf("https://api.%s.onelogin.com", (*profile).Region)
			}
			oneloginClient, err := client.NewClient(&clientConfig)
			if err != nil {
				log.Println("[Warning] Unable to connect to OneLogin!", err)
			}
			// initalize other clients to inject into respective importable services here
			importables := map[string]tfimportables.Importable{
				"onelogin_users":         tfimportables.OneloginUsersImportable{Service: oneloginClient.Services.UsersV2, SearchID: *searchId},
				"onelogin_apps":          tfimportables.OneloginAppsImportable{Service: oneloginClient.Services.AppsV2, SearchID: *searchId},
				"onelogin_saml_apps":     tfimportables.OneloginAppsImportable{Service: oneloginClient.Services.AppsV2, SearchID: *searchId, AppType: "onelogin_saml_apps"},
				"onelogin_oidc_apps":     tfimportables.OneloginAppsImportable{Service: oneloginClient.Services.AppsV2, SearchID: *searchId, AppType: "onelogin_oidc_apps"},
				"onelogin_user_mappings": tfimportables.OneloginUserMappingsImportable{Service: oneloginClient.Services.UserMappingsV2, SearchID: *searchId},
			}
			importable, ok := importables[strings.ToLower(args[0])]

			if !ok {
				availableImportables := make([]string, 0, len(importables))
				for k := range importables {
					availableImportables = append(availableImportables, fmt.Sprintf("%s", k))
				}
				log.Fatalf("Available resources: %s", availableImportables)
			}

			tfImport(importable, args, *autoApprove)
		},
	}
	autoApprove = tfImportCommand.Flags().Bool("auto_approve", false, "Skip confirmation of resource import")
	searchId = tfImportCommand.Flags().String("id", "", "Import one resource by id")
	rootCmd.AddCommand(tfImportCommand)
}

func tfImport(importable tfimportables.Importable, args []string, autoApprove bool) {
	planFile, err := os.OpenFile(filepath.Join("main.tf"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to open main.tf ", err)
	}

	newResourceDefinitions, newProviderDefinitions := tfimport.FilterExistingDefinitions(planFile, importable)
	if len(newResourceDefinitions) == 0 {
		fmt.Println("No new resources to import from remote")
		planFile.Close()
		os.Exit(0)
	}

	if autoApprove == false {
		fmt.Printf("This will import %d resources. Do you want to continue? (y/n): ", len(newResourceDefinitions))
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		text := strings.ToLower(input.Text())
		if text != "y" && text != "yes" {
			fmt.Printf("User aborted operation!")
			if err := planFile.Close(); err != nil {
				fmt.Println("Problem writing file", err)
			}
			os.Exit(0)
		}
	}

	if err := tfimport.WriteHCLDefinitionHeaders(newResourceDefinitions, newProviderDefinitions, planFile); err != nil {
		planFile.Close()
		log.Fatal("Problem creating import file", err)
	}

	log.Println("Initializing Terraform with 'terraform init'...")
	// #nosec G204
	if err := exec.Command("terraform", "init").Run(); err != nil {
		if err := planFile.Close(); err != nil {
			log.Fatal("Problem writing to main.tf", err)
		}
		log.Fatal("Problem executing terraform init", err)
	}

	for i, resourceDefinition := range newResourceDefinitions {
		arg1 := fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)
		pos := strings.Index(arg1, "-")
		id := arg1[pos+1 : len(arg1)]
		// #nosec G204
		cmd := exec.Command("terraform", "import", arg1, id)
		log.Printf("Importing resource %d", i+1)
		if err := cmd.Run(); err != nil {
			log.Fatal("Problem executing terraform import", cmd.Args, err)
		}
	}

	// grab the state from tfstate
	state := stateparser.State{}
	log.Println("Collecting State from tfstate File")
	data, err := ioutil.ReadFile(filepath.Join("terraform.tfstate"))
	if err != nil {
		planFile.Close()
		log.Fatalln("Unable to Read tfstate", err)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		planFile.Close()
		log.Fatalln("Unable to Translate tfstate in Memory", err)
	}

	buffer := stateparser.ConvertTFStateToHCL(state, importable)

	// go to the start of main.tf and overwrite whole file
	planFile.Seek(0, 0)
	_, err = planFile.Write(buffer)
	if err != nil {
		planFile.Close()
		fmt.Println("ERROR Writing Final main.tf", err)
	}

	if err := planFile.Close(); err != nil {
		fmt.Println("Problem writing file", err)
	}
}
