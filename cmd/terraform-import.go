package cmd

import (
	"fmt"
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

type TerraformResource struct {
	ImportCommand *exec.Cmd   // the terraform import command for the resource
	Content       []byte      // the []byte representation of main.tf contents
	Name          string      // app.Name-app.ID with all special chars removed
	Type          string      // terraform resource type (e.g. onelogin_saml_apps)
	Resource      interface{} // the resource itself
}

/// what if we just peel apart the resource and parse sub-resources separately

func (tfResource *TerraformResource) CreateResourceFileOutput() {
	head := fmt.Sprintf("resource %s %s {\n", tfResource.Type, tfResource.Name)
	tfResource.Content = append(tfResource.Content, []byte(head)...)
	tfResource.Content = append(tfResource.Content, resourceBaseToHCL(tfResource.Resource)...)
	tfResource.Content = append(tfResource.Content, fmt.Sprintf("}\n\n")...)
}

func (tfResource *TerraformResource) CreateTerraformImportCommand() {
	arg1 := fmt.Sprintf("%s.%s", tfResource.Type, tfResource.Name)
	pos := strings.Index(arg1, "-")
	id := arg1[pos+1 : len(arg1)]
	tfResource.ImportCommand = exec.Command("terraform", "import", arg1, id)
}

func terraformImport(cmd *cobra.Command, args []string) {
	fmt.Println("Terraform Import!")

	allApps := getAllApps()

	f := setUpTerraformFile()
	defer f.Close()

	var resourcesToBootstrap []TerraformResource
	for _, app := range allApps {
		resource := createResourceRepresentation(app)
		resourcesToBootstrap = append(resourcesToBootstrap, resource)
	}
	writeInitialMainTF(f, resourcesToBootstrap)
	importTFState(resourcesToBootstrap)
	// read the .tfstate file as json and update each resource in main.tf

}

func importTFState(resources []TerraformResource) {
	exec.Command("terraform", "init").Run()
	for _, resource := range resources {
		if err := resource.ImportCommand.Run(); err != nil {
			fmt.Println("ERROR IN TF", err)
		}
	}
}

func writeInitialMainTF(f *os.File, resources []TerraformResource) {
	var contents []byte
	for _, resource := range resources {
		contents = append(contents, resource.Content...)
	}
	f.Write(contents)
}

func createResourceRepresentation(app models.App) TerraformResource {
	resource := TerraformResource{
		Resource: app,
	}
	// todo: this would be better off as a config file or redis lookup
	switch *app.AuthMethod {
	case 1:
		resource.Type = "onelogin_oidc_apps"
	case 2:
		resource.Type = "onelogin_saml_apps"
	default:
		resource.Type = "onelogin_apps"
	}
	resource.Name = fmt.Sprintf("%s-%d", toSnakeCase(replaceSpecialChar(*app.Name, "")), *app.ID)
	resource.CreateResourceFileOutput()
	resource.CreateTerraformImportCommand()
	return resource
}

func resourceBaseToHCL(input interface{}) []byte {
	var out []byte

	tp := reflect.TypeOf(input)
	vl := reflect.ValueOf(input)

	for i := 0; i < tp.NumField(); i++ {
		q := vl.Field(i)
		if !q.IsZero() {
			switch q.Kind() {
			case reflect.Ptr:
				switch q.Elem().Kind() {
				case reflect.String:
					out = append(out, []byte(fmt.Sprintf("\t%s = \"%s\"\n", toSnakeCase(tp.Field(i).Name), q.Elem()))...)
				case reflect.Bool, reflect.Int, reflect.Int32, reflect.Int64:
					out = append(out, []byte(fmt.Sprintf("\t%s = %v\n", toSnakeCase(tp.Field(i).Name), q.Elem()))...)
				default:
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

func setUpTerraformFile() *os.File {
	path, _ := os.Getwd()
	p := filepath.Join(path, "/main.tf")
	f, err := os.Create(p)
	if err != nil {
		log.Fatal("ERROR CREATING FILE", err)
	}
	f.Write([]byte("provider onelogin {}\n\n"))
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
