package terraform

import (
	"bufio"
	"fmt"
	"io"
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
	ImportResourceDefinitionsFromRemote() []ResourceDefinition
}

// ResourceDefinition represents the resource to be imported
type ResourceDefinition struct {
	Content  []byte
	Provider string
	Name     string
	Type     string
}

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

	newResourceDefinitions := importable.ImportResourceDefinitionsFromRemote()
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
	createTFStateFromRemote(newResourceDefinitions)
}

// UpdateMainTFFromState reads .tfstate and updates the main.tf as if the tfstate was
// created with the ensuing main.tf file
func UpdateMainTFFromState() {
	path, _ := os.Getwd()
	p := filepath.Join(path, ("/main.tf"))
	f, err := os.OpenFile(p, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error Creating Output main.tf ", err)
	}
	defer f.Close()

	state := State{}
	state.Initialize()

	buffer := convertTFStateToHCL(state)

	_, err = f.Write(buffer)
	if err != nil {
		fmt.Println("ERROR Writing Final main.tf", err)
	}
}

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
			resource.Content = append(resource.Content, resourceBaseToHCL(instance.Data, 1)...)
			resource.Content = append(resource.Content, []byte("}\n\n")...)
		}
		buffer = append(buffer, resource.Content...)
	}
	return buffer
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

func createTFStateFromRemote(resourceDefinitions []ResourceDefinition) {
	log.Println("Initializing Terraform with 'terraform init'...")
	if err := exec.Command("terraform", "init").Run(); err != nil {
		log.Fatal("Problem executing terraform init", err)
	}
	for i, resourceDefinition := range resourceDefinitions {
		log.Printf("Importing resource %d of %d", i+1, len(resourceDefinitions))
		arg1 := fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)
		pos := strings.Index(arg1, "-")
		id := arg1[pos+1 : len(arg1)]
		cmd := exec.Command("terraform", "import", arg1, id)
		if err := cmd.Run(); err != nil {
			log.Fatal("Problem executing terraform import", cmd.Args, err)
		}
	}
}
