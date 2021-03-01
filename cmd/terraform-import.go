package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onelogin/onelogin/clients"
	"github.com/onelogin/onelogin/profiles"
	tfimport "github.com/onelogin/onelogin/terraform/import"
	tfimportables "github.com/onelogin/onelogin/terraform/importables"
	stateparser "github.com/onelogin/onelogin/terraform/state_parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var (
		autoApprove   *bool
		outFile       *string
		searchID      *string
		clientConfigs clients.ClientConfigs
	)
	var tfImportCommand = &cobra.Command{
		Use:   "terraform-import",
		Short: `Import resources to local Terraform state.`,
		Long: `Uses Terraform Import to collect resources from a remote and
		create new .tfstate and .tf files so you can begin managing existing resources with Terraform.
		Available Imports:
			onelogin_apps          => onelogin all apps
			onelogin_saml_apps     => onelogin SAML apps only
			onelogin_oidc_apps     => onelogin OIDC apps only
			onelogin_user_mappings => onelogin user mappings
			onelogin_users         => onelogin users
			onelogin_roles         => onelogin roles
			aws_iam_user           => aws users`,
		Args: cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			configFile, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR, 0600)
			if err != nil {
				configFile.Close()
				log.Println("Unable to open profiles file. Falling back to Environment Variables", err)
			}
			profileService := profiles.ProfileService{
				Repository: profiles.FileRepository{
					StorageMedia: configFile,
				},
			}
			profile := profileService.GetActive()
			clientConfigs = clients.ClientConfigs{
				AwsRegion: os.Getenv("AWS_REGION"),
			}
			if profile == nil {
				fmt.Println("No active profile detected. Authenticating with environment variables")
				clientConfigs.OneLoginClientID = os.Getenv("ONELOGIN_CLIENT_ID")
				clientConfigs.OneLoginClientSecret = os.Getenv("ONELOGIN_CLIENT_SECRET")
				clientConfigs.OneLoginURL = os.Getenv("ONELOGIN_OAPI_URL")
			} else {
				fmt.Println("Using profile", (*profile).Name)
				clientConfigs.OneLoginClientID = (*profile).ClientID
				clientConfigs.OneLoginClientSecret = (*profile).ClientSecret
				clientConfigs.OneLoginURL = fmt.Sprintf("https://api.%s.onelogin.com", (*profile).Region)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			tfImport(args, clientConfigs, *autoApprove, searchID, *outFile)
		},
	}
	autoApprove = tfImportCommand.Flags().BoolP("auto_approve", "a", false, "Skip confirmation of resource import")
	outFile = tfImportCommand.Flags().StringP("output", "o", "", "Output filename")
	searchID = tfImportCommand.Flags().StringP("id", "i", "", "Import one resource by id")
	rootCmd.AddCommand(tfImportCommand)
}

func tfImport(args []string, clientConfigs clients.ClientConfigs, autoApprove bool, searchID *string, outFile string) {
	sourceName := args[0]
	workingDir, _ := os.Getwd()
	if outFile == "" {
		outFile = fmt.Sprintf("%s.tf", strings.Split(sourceName, "_")[0])
	}
	// #nosec G304 forcing the file to be created in the working directory
	planFile, err := os.OpenFile(filepath.Join(workingDir, outFile), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("Unable to create desired tf file ", err)
	}

	pfReader, err := ioutil.ReadAll(planFile)
	if err != nil {
		planFile.Close()
		log.Fatalln("Unable to read from tf file ", err)
	}

	clientList := clients.New(clientConfigs)
	importables := tfimportables.New(clientList)
	importable := importables.GetImportable(strings.ToLower(args[0]))

	pfReader1 := bytes.NewReader(pfReader)
	resourceDefinitionsFromRemote := importable.ImportFromRemote(searchID)
	newResourceDefinitions, newProviderDefinitions := tfimport.DetermineNewResourcesAndProviders(pfReader1, resourceDefinitionsFromRemote)
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

	pfReader2 := bytes.NewReader(pfReader)
	newHCL := tfimport.AddNewProvidersAndResourceHCL(pfReader2, newResourceDefinitions, newProviderDefinitions)

	planFile.Seek(0, 0)
	if _, err := planFile.Write([]byte(newHCL)); err != nil {
		planFile.Close()
		log.Fatal("Problem creating import file", err)
	}

	log.Println("Initializing Terraform with 'terraform init'...")
	// #nosec G204 running prescribed terraform command
	if err := exec.Command("terraform", "init").Run(); err != nil {
		if err := planFile.Close(); err != nil {
			log.Fatal("Problem writing to tf file ", err)
		}
		log.Fatal("Problem executing terraform init", err)
	}

	for i, resourceDefinition := range newResourceDefinitions {
		resourceName := fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)
		id := resourceDefinition.ImportID
		// #nosec G204 running prescribed terraform command
		cmd := exec.Command("terraform", "import", resourceName, id)
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

	buffer := stateparser.ConvertTFStateToHCL(state, importables)

	// go to the start of main.tf and overwrite whole file
	planFile.Seek(0, 0)
	_, err = planFile.Write(buffer)
	if err != nil {
		planFile.Close()
		fmt.Println("ERROR Writing Final tf file ", err)
	}

	if err := planFile.Close(); err != nil {
		fmt.Println("Problem writing file", err)
	}
}
