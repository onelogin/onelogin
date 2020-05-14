package terraform

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/onelogin/onelogin-cli/utils"
)

type Importable interface {
	ImportFromRemote() []ResourceDefinition
}

// ResourceDefinition represents the resource to be imported
type ResourceDefinition struct {
	Content  []byte
	Provider string
	Name     string
	Type     string
}

// ImportTFStateFromRemote writes the resource resourceDefinitions to main.tf and calls each
// resource's terraform import command to update tfstate
func ImportTFStateFromRemote(importable Importable) {
	path, _ := os.Getwd()
	p := filepath.Join(path, ("/main.tf"))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("Unable to open main.tf ", err)
	}
	defer f.Close()

	newResourceDefinitions := importable.ImportFromRemote()
	existingDefinitions := collectTerraformDefinitionsFromFile(f)
	newResourceDefinitions, newProviderDefinitions := filterExistingDefinitions(existingDefinitions, newResourceDefinitions)

	if len(newResourceDefinitions) == 0 {
		fmt.Println("No new resources to import from remote")
		os.Exit(0)
	}

	fmt.Printf("This will import %d resources. Do you want to continue? (y/n): ", len(newResourceDefinitions))
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	text := strings.ToLower(input.Text())
	if text != "y" && text != "yes" {
		fmt.Printf("User aborted operation!")
		os.Exit(0)
	}

	appendDefinitionsToMainTF(f, newResourceDefinitions, newProviderDefinitions)
	importTFStateFromRemote(newResourceDefinitions)
}

// UpdateMainTFFromState reads .tfstate and updates the main.tf as if the tfstate was
// created with the ensuing main.tf file
func UpdateMainTFFromState() {
	path, _ := os.Getwd()
	f, err := os.OpenFile(filepath.Join(path, ("/main.tf")), os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error Creating Output main.tf ", err)
	}
	defer f.Close()

	state := State{}
	log.Println("Collecting State from tfstate File")
	data, err := ioutil.ReadFile(filepath.Join(path, "/terraform.tfstate"))
	if err != nil {
		log.Fatal("Unable to Read tfstate")
	}

	if err := json.Unmarshal(data, &state); err != nil {
		log.Fatal("Unable to Translate tfstate in Memory")
	}

	buffer := convertTFStateToHCL(state)

	_, err = f.Write(buffer)
	if err != nil {
		fmt.Println("ERROR Writing Final main.tf", err)
	}
}

// compares incoming resources from remote to what is already defined in the main.tf
// file to prevent duplicate definitions which breaks terraform import
func filterExistingDefinitions(countsFromFile map[string]map[string]int, resourceDefinitions []ResourceDefinition) ([]ResourceDefinition, []string) {
	uniqueResourceDefinitions := []ResourceDefinition{}
	uniqueProviders := []string{}
	providerMap := map[string]int{}

	for _, resourceDefinition := range resourceDefinitions {
		providerMap[resourceDefinition.Provider]++
		if countsFromFile["resource"][fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)] == 0 {
			uniqueResourceDefinitions = append(uniqueResourceDefinitions, resourceDefinition)
		}
	}

	for provider := range providerMap {
		if countsFromFile["provider"][provider] == 0 {
			uniqueProviders = append(uniqueProviders, provider)
		}
	}

	return uniqueResourceDefinitions, uniqueProviders
}

// Scans the existing main.tf file for any existing resource and provider definitions
// and returns a count of unique definitions
func collectTerraformDefinitionsFromFile(f io.Reader) map[string]map[string]int {
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
				var definitionKey string
				if regexName == "provider" {
					definitionKey = fmt.Sprintf("%s", subStr[len(subStr)-1])
				}
				if regexName == "resource" {
					definitionKey = fmt.Sprintf("%s.%s", subStr[len(subStr)-2], subStr[len(subStr)-1])
				}
				collection[regexName][definitionKey]++
			}
		}
	}
	return collection
}

// takes the tfstate representations formats them as HCL and writes them to a bytes buffer
// so it can be flushed into main.tf
func convertTFStateToHCL(state State) []byte {
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
			resource.Content = append(resource.Content, convertToHCLByteSlice(instance.Data, 1)...)
			resource.Content = append(resource.Content, []byte("}\n\n")...)
		}
		buffer = append(buffer, resource.Content...)
	}
	return buffer
}

// recursively converts a chunk of data from it's struct representation to its HCL representation
// and appends the "line" to a bytes buffer.
func convertToHCLByteSlice(input interface{}, indentLevel int) []byte {
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
					out = append(out, convertToHCLByteSlice(field.Index(j).Interface(), indentLevel+1)...)
					out = append(out, []byte("\t}\n")...)
				}
			}
		}
	}
	return out
}

// in preparation for terraform import, appends empty resource definitions to the existing main.tf file
func appendDefinitionsToMainTF(f io.ReadWriter, resourceDefinitions []ResourceDefinition, providerDefinitions []string) {
	var buffer []byte
	for _, newProvider := range providerDefinitions {
		buffer = append(buffer, []byte(fmt.Sprintf("provider %s {\n\talias = \"%s\"\n}\n\n", newProvider, newProvider))...)
	}

	for _, resourceDefinition := range resourceDefinitions {
		resourceDefinition.Content = append(resourceDefinition.Content, []byte(fmt.Sprintf("resource %s %s {}\n", resourceDefinition.Type, resourceDefinition.Name))...)
		buffer = append(buffer, resourceDefinition.Content...)
	}
	if _, err := f.Write(buffer); err != nil {
		log.Fatal("Problem creating import file", err)
	}
}

// loops over the resources to import and calls terraform import with the required resoruce arguments
func importTFStateFromRemote(resourceDefinitions []ResourceDefinition) {
	log.Println("Initializing Terraform with 'terraform init'...")
	if err := exec.Command("terraform", "init").Run(); err != nil {
		log.Fatal("Problem executing terraform init", err)
	}
	for i, resourceDefinition := range resourceDefinitions {
		arg1 := fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)
		pos := strings.Index(arg1, "-")
		id := arg1[pos+1 : len(arg1)]
		cmd := exec.Command("terraform", "import", arg1, id)
		log.Printf("Importing resource %d", i+1)
		if err := cmd.Run(); err != nil {
			log.Fatal("Problem executing terraform import", cmd.Args, err)
		}
	}
}
