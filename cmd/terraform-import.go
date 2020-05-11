package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/onelogin/onelogin-go-sdk/pkg/client"
	"github.com/onelogin/onelogin-go-sdk/pkg/models"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "terraform-import",
		Short: "Import apps to local Terraform state",
		Long: `Uses Terraform Import to collect existing OneLogin apps
    and create new .tfstate and .tf files so you can begin managing
    existing resources with Terraform.`,
		Run: terraformImport,
	})
}

type TerraformPlaceholder struct {
	ImportCommand *exec.Cmd // the terraform import command for the resource
	Content       []byte    // the []byte representation of main.tf contents
	Name          string    // app.Name-app.ID with all special chars removed
	Type          string    // terraform resource type (e.g. onelogin_saml_apps)
}

func (placeholder *TerraformPlaceholder) InitializeTerraformImport() {
	placeholder.Content = append(placeholder.Content, []byte(fmt.Sprintf("resource %s %s {}\n", placeholder.Type, placeholder.Name))...)
	arg1 := fmt.Sprintf("%s.%s", placeholder.Type, placeholder.Name)
	pos := strings.Index(arg1, "-")
	id := arg1[pos+1 : len(arg1)]
	placeholder.ImportCommand = exec.Command("terraform", "import", arg1, id)
}

func terraformImport(cmd *cobra.Command, args []string) {
	availableImportArgs := []string{"apps"}
	fmt.Println("Terraform Import!")

	if len(args) == 0 {
		fmt.Println("Must specify resource to import!")
		fmt.Println("Available resources:", availableImportArgs)
		os.Exit(1)
	}
	fmt.Println("Collecting Apps from OneLogin...")

	allApps := getAllApps()

	fmt.Printf("This will import %d apps. Do you want to continue? (y/n): ", len(allApps))
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	ans := strings.ToLower(input.Text())
	if ans != "y" && ans != "yes" {
		os.Exit(0)
	}

	f := setUpTerraformFile("main")
	defer f.Close()

	placeholders := make([]TerraformPlaceholder, len(allApps))

	for i, app := range allApps {
		placeholder := TerraformPlaceholder{}
		switch *app.AuthMethod {
		case 1, 8:
			placeholder.Type = "onelogin_oidc_apps"
		case 2:
			placeholder.Type = "onelogin_saml_apps"
		default:
			placeholder.Type = "onelogin_apps"
		}
		placeholder.Name = fmt.Sprintf("%s-%d", toSnakeCase(replaceSpecialChar(*app.Name, "")), *app.ID)
		placeholder.InitializeTerraformImport()
		placeholders[i] = placeholder
	}
	state := importTFState(f, placeholders)
	writeFinalMainTF(f, state.Resources)
}

type InstanceData struct {
	AllowAssumedSignin *bool                     `json:"allow_assumed_signin"`
	ConnectorID        *int                      `json:"connector_id"`
	Description        *string                   `json:"description"`
	Name               *string                   `json:"name"`
	Notes              *string                   `json:"notes"`
	Visible            *bool                     `json:"visible"`
	Provisioning       []models.AppProvisioning  `json:"provisioning"`
	Parameters         []models.AppParameters    `json:"parameters"`
	Configuration      []models.AppConfiguration `json:"configuration"`
}

type ResourceInstance struct {
	Data InstanceData `json:"attributes"`
}

type TerraformResource struct {
	Content   []byte
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Instances []ResourceInstance `json:"instances"`
}

type TerraformState struct {
	Resources []TerraformResource `json:"resources"`
}

func writeFinalMainTF(f *os.File, resources []TerraformResource) {
	log.Println("Assembling main.tf...")
	buffer := []byte("provider onelogin {}\n\n")
	for _, resource := range resources {
		for _, instance := range resource.Instances {
			resource.Content = append(resource.Content, []byte(fmt.Sprintf("resource %s %s {\n", resource.Type, resource.Name))...)
			resource.Content = append(resource.Content, resourceBaseToHCL(instance.Data, 1)...)
			resource.Content = append(resource.Content, []byte("}\n\n")...)
		}
		buffer = append(buffer, resource.Content...)
	}
	f.WriteAt(buffer, 0)
}

func readTFStateJSON() TerraformState {
	log.Println("Collecting State from State File")

	path, _ := os.Getwd()
	p := filepath.Join(path, "/terraform.tfstate")
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal("Unable to Read tfstate")
	}
	v := TerraformState{}

	if err := json.Unmarshal(data, &v); err != nil {
		log.Fatal("Unable to Translate tfstate in Memory")
	}
	return v
}

func importTFState(f *os.File, placeholders []TerraformPlaceholder) TerraformState {
	log.Println("Creating Terraform Import File...")
	buffer := []byte("provider onelogin {}\n\n")
	for _, placeholder := range placeholders {
		buffer = append(buffer, placeholder.Content...)
	}
	if _, err := f.Write(buffer); err != nil {
		log.Fatal("Problem creating import file", err)
	}

	log.Println("Initializing Terraform with 'terraform init'...")
	exec.Command("terraform", "init").Run()
	for i, placeholder := range placeholders {
		log.Printf("Importing resource %d of %d", i+1, len(placeholders))
		if err := placeholder.ImportCommand.Run(); err != nil {
			log.Fatal("Problem executing terraform import", placeholder.ImportCommand.Args, err)
		}
	}
	return readTFStateJSON()
}

func resourceBaseToHCL(input interface{}, indentLevel int) []byte {
	var out []byte

	tp := reflect.TypeOf(input)
	vl := reflect.ValueOf(input)

	for i := 0; i < tp.NumField(); i++ {
		line := make([]byte, indentLevel)
		for i := 0; i < indentLevel; i++ {
			line[i] = byte('\t')
		}

		field := vl.Field(i)
		if !field.IsZero() {
			switch field.Kind() {
			case reflect.Ptr:
				switch field.Elem().Kind() {
				case reflect.String:
					line = append(line, []byte(fmt.Sprintf("%s = \"%s\"\n", toSnakeCase(tp.Field(i).Name), field.Elem()))...)
					out = append(out, line...)
				case reflect.Bool, reflect.Int, reflect.Int32, reflect.Int64:
					line = append(line, []byte(fmt.Sprintf("%s = %v\n", toSnakeCase(tp.Field(i).Name), field.Elem()))...)
					out = append(out, line...)
				default:
					fmt.Println("Unable to Determine Type")
				}
			case reflect.Array, reflect.Slice:
				for j := 0; j < field.Len(); j++ {
					out = append(out, []byte(strings.ToLower(fmt.Sprintf("\n\t%s {\n", tp.Field(i).Name)))...)
					out = append(out, resourceBaseToHCL(field.Index(j).Interface(), indentLevel+1)...)
					out = append(out, []byte("\t}\n")...)
				}
			}
		}
	}

	return out
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
		log.Fatal("error retrieving apps", err)
	}
	return allApps
}

func setUpTerraformFile(name string) *os.File {
	path, _ := os.Getwd()
	p := filepath.Join(path, fmt.Sprintf("/%s.tf", name))
	f, err := os.Create(p)
	if err != nil {
		log.Fatal("ERROR CREATING FILE", err)
	}
	return f
}

func replaceSpecialChar(str string, rep string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9\\s]+")
	return reg.ReplaceAllString(str, rep)
}

func toSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")
	matchAllSpaces := regexp.MustCompile("(\\s)")
	cleanUpHack := regexp.MustCompile("i_ds")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	snake = matchAllSpaces.ReplaceAllString(snake, "_")
	snake = strings.ToLower(snake)
	snake = cleanUpHack.ReplaceAllString(snake, "ids")

	return snake
}
