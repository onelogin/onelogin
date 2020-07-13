package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin/profiles"
	"github.com/onelogin/onelogin/terraform/import"

	"github.com/onelogin/onelogin/terraform/importables"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var (
		autoApprove    *bool
		searchId       *string
		profileService profiles.ProfileService
	)
	var convertCommand = &cobra.Command{
		Use:   "convert [source] [destination]",
		Short: `Converts resoruces from source format to destination format.`,
		Long: `Uses Terraform to collect resources from a source,
		changes them into destination-compatible resources,
		and uses Terraform to upload them to the destination remote.
		Available Converstions:
			okta_apps onelogin_apps => convert okta apps to onelogin apps`,
		Args: cobra.MinimumNArgs(2),
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

			_, oktaClient, err := okta.NewClient(context.TODO(), okta.WithOrgUrl(fmt.Sprintf("https://%s.%s", os.Getenv("OKTA_ORG_NAME"), os.Getenv("OKTA_BASE_URL"))), okta.WithToken(os.Getenv("OKTA_API_TOKEN")))
			if err != nil {
				log.Println("[Warning] Unable to connect to Okta!", err)
			}

			availableSources := map[string]tfimportables.Importable{
				"okta_apps": tfimportables.OktaAppsImportable{Service: oktaClient.Application, SearchID: *searchId},
			}
			source, ok := availableSources[strings.ToLower(args[0])]

			if !ok {
				log.Fatalln("Available conversions: okta_apps => onelogin_apps")
			}

			convert(source, args, *autoApprove)
		},
	}
	autoApprove = convertCommand.Flags().Bool("auto_approve", false, "Skip confirmation of resource import")
	searchId = convertCommand.Flags().String("id", "", "Import one resource by id")
	rootCmd.AddCommand(convertCommand)
}

func convert(importable tfimportables.Importable, args []string, autoApprove bool) {
	// create a main.tf file
	p := filepath.Join("main.tf")
	planFile, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to open main.tf ", err)
	}

	// call out for source resoruces
	resourceDefinitions := importable.ImportFromRemote()
	providerDefinitions := []string{"okta"} // good enough for hackathon lol
	// ask for permission
	if autoApprove == false {
		fmt.Printf("This will import %d resources. Do you want to continue? (y/n): ", len(resourceDefinitions))
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

	// adds resource headers to main.tf e.g. resource okta_saml_apps okta_saml_apps-1234 {}
	if err := tfimport.WriteHCLDefinitionHeaders(resourceDefinitions, providerDefinitions, planFile); err != nil {
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

	// import each resource with terraform import
	for i, resourceDefinition := range resourceDefinitions {
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
	state, err := tfimport.CollectState()
	if err != nil {
		planFile.Close()
		log.Fatalln("Unable to collect state from tfstate")
	}

	buffer := tfimport.ConvertTFStateToDestinationHCL(state, importable)

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
