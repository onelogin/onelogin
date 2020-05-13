package terraform

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/onelogin/onelogin-cli/utils"
	"github.com/onelogin/onelogin-go-sdk/pkg/models"
)

// ResourceDefinition represents the resource to be imported
type ResourceDefinition struct {
	ImportCommand *exec.Cmd
	Content       []byte
	Provider      string
	Name          string
	Type          string
}

// PrepareTFImport creates the representation of the resource to be imported
// with identifying information and import commands
func (resourceDefinition *ResourceDefinition) PrepareTFImport() {
	resourceDefinition.Content = append(resourceDefinition.Content, []byte(fmt.Sprintf("resource %s %s {}\n", resourceDefinition.Type, resourceDefinition.Name))...)
	arg1 := fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)
	pos := strings.Index(arg1, "-")
	id := arg1[pos+1 : len(arg1)]
	resourceDefinition.ImportCommand = exec.Command("terraform", "import", arg1, id)
}

// State in memory representation of tfstate
type State struct {
	Resources []Resource `json:"resources"`
}

// Resource represents each resource type's list of resources
type Resource struct {
	Content   []byte
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Provider  string             `json:"provider"`
	Instances []resourceInstance `json:"instances"`
}

type resourceInstance struct {
	Data instanceData `json:"attributes"`
}

type instanceData struct {
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

func filterExistingResourceDefinitions(countsFromFile map[string]int, resourceDefinitions []ResourceDefinition) []ResourceDefinition {
	var deduped []ResourceDefinition
	for _, resourceDefinition := range resourceDefinitions {
		if countsFromFile[fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)] == 0 {
			deduped = append(deduped, resourceDefinition)
		}
	}
	return deduped
}

func filterExistingProviderDefinitions(countsFromFile map[string]int, resourceDefinitions []ResourceDefinition) []string {
	var deduped []string
	incomingProviders := map[string]int{}
	for _, resourceDefinition := range resourceDefinitions {
		incomingProviders[resourceDefinition.Provider]++
	}
	for incomingProvider := range incomingProviders {
		if countsFromFile[fmt.Sprintf("%s.%s", incomingProvider, incomingProvider)] == 0 {
			deduped = append(deduped, incomingProvider)
		}
	}
	return deduped
}

// Scans the existing main.tf file for any existing resource and provider definitions
// and returns a count of unique definitions
func collectTerraformDefinitionsFromFile(f *os.File) map[string]map[string]int {
	searchCriteria := map[string]*regexp.Regexp{
		"provider": regexp.MustCompile(`(\w*provider\w*)\s(([a-zA-Z\_]*))\s\{`),
		"resource": regexp.MustCompile(`(\w*resource\w*)\s([a-zA-Z\_\-]*)\s([a-zA-Z\_\-]*[0-9]*)\s?\{`),
	}
	collection := make(map[string]map[string]int)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		for regexName, r := range searchCriteria {
			if collection[regexName] == nil {
				collection[regexName] = make(map[string]int)
			}
			subStr := r.FindStringSubmatch(t)
			if len(subStr) > 0 {
				definitionKey := fmt.Sprintf("%s.%s", subStr[len(subStr)-2], subStr[len(subStr)-1])
				collection[regexName][definitionKey]++
			}
		}
	}
	return collection
}

// ImportTFStateFromRemote writes the resource resourceDefinitions to main.tf and calls each
// resource's terraform import command to update tfstate
func ImportTFStateFromRemote(resourceDefinitions []ResourceDefinition) {
	path, _ := os.Getwd()
	p := filepath.Join(path, ("/main.tf"))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println("Unable to open main.tf ", err)
	}
	defer f.Close()

	existingDefinitions := collectTerraformDefinitionsFromFile(f)

	resourceDefinitions = filterExistingResourceDefinitions(existingDefinitions["resource"], resourceDefinitions)
	if len(resourceDefinitions) == 0 {
		fmt.Println("No new resources to import from remote")
		os.Exit(0)
	}

	newProviders := filterExistingProviderDefinitions(existingDefinitions["provider"], resourceDefinitions)

	fmt.Printf("This will import %d resources. Do you want to continue? (y/n): ", len(resourceDefinitions))
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	text := strings.ToLower(input.Text())
	if text != "y" && text != "yes" {
		fmt.Printf("User aborted operation!")
		os.Exit(0)
	}

	var buffer []byte
	for _, newProvider := range newProviders {
		buffer = append(buffer, []byte(fmt.Sprintf("provider %s {\n\talias = \"%s\"\n}\n\n", newProvider, newProvider))...)
	}

	for _, resourceDefinition := range resourceDefinitions {
		buffer = append(buffer, resourceDefinition.Content...)
	}

	if _, err := f.Write(buffer); err != nil {
		log.Fatal("Problem creating import file", err)
	}
	log.Println("Initializing Terraform with 'terraform init'...")
	if err := exec.Command("terraform", "init").Run(); err != nil {
		log.Fatal("Problem executing terraform init", err)
	}

	for i, resourceDefinition := range resourceDefinitions {
		log.Printf("Importing resource %d of %d", i+1, len(resourceDefinitions))
		if err := resourceDefinition.ImportCommand.Run(); err != nil {
			log.Fatal("Problem executing terraform import", resourceDefinition.ImportCommand.Args, err)
		}
	}
}

// UpdateMainTFFromState reads .tfstate and updates the main.tf as if the tfstate was
// created with the ensuing main.tf file
func UpdateMainTFFromState() {
	state := readTFStateJSON()

	path, _ := os.Getwd()
	p := filepath.Join(path, ("/main.tf"))
	f, err := os.OpenFile(p, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error Creating Output main.tf ", err)
	}
	defer f.Close()

	var buffer []byte
	knownProviders := map[string]int{}

	log.Println("Assembling main.tf...")

	for _, resource := range state.Resources {
		providerDefinition := strings.Replace(resource.Provider, "provider.", "", 1)
		if knownProviders[providerDefinition] == 0 {
			knownProviders[providerDefinition]++
			buffer = append(buffer, []byte(fmt.Sprintf("provider %s {\n\talias = \"%s\"\n}\n\n", providerDefinition, providerDefinition))...)
		}
		for _, instance := range resource.Instances {
			resource.Content = append(resource.Content, []byte(fmt.Sprintf("resource %s %s {\n", resource.Type, resource.Name))...)
			resource.Content = append(resource.Content, []byte(fmt.Sprintf("\tprovider = %s\n", providerDefinition))...)
			resource.Content = append(resource.Content, resourceBaseToHCL(instance.Data, 1)...)
			resource.Content = append(resource.Content, []byte("}\n\n")...)
		}
		buffer = append(buffer, resource.Content...)
	}
	_, err = f.Write(buffer)
	if err != nil {
		fmt.Println("ERROR Writing Final main.tf", err)
	}
}

func readTFStateJSON() State {
	log.Println("Collecting State from State File")

	path, _ := os.Getwd()
	p := filepath.Join(path, "/terraform.tfstate")
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal("Unable to Read tfstate")
	}
	v := State{}

	if err := json.Unmarshal(data, &v); err != nil {
		log.Fatal("Unable to Translate tfstate in Memory")
	}
	return v
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
					line = append(line, []byte(fmt.Sprintf("%s = \"%s\"\n", utils.ToSnakeCase(tp.Field(i).Name), field.Elem()))...)
					out = append(out, line...)
				case reflect.Bool, reflect.Int, reflect.Int32, reflect.Int64:
					line = append(line, []byte(fmt.Sprintf("%s = %v\n", utils.ToSnakeCase(tp.Field(i).Name), field.Elem()))...)
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
