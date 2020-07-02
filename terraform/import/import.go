package tfimport

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
	"github.com/onelogin/onelogin/terraform/importables"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

// State is the in memory representation of tfstate.
type State struct {
	Resources []StateResource `json:"resources"`
}

// Terraform resource representation
type StateResource struct {
	Content   []byte
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Provider  string             `json:"provider"`
	Instances []ResourceInstance `json:"instances"`
}

// An instance of a particular resource without the terraform information
type ResourceInstance struct {
	Data interface{} `json:"attributes"`
}

func CollectState() (State, error) {
	state := State{}
	log.Println("Collecting State from tfstate File")
	data, err := ioutil.ReadFile(filepath.Join("terraform.tfstate"))
	if err != nil {
		log.Println(err)
		return state, errors.New("Unable to Read tfstate")
	}

	if err := json.Unmarshal(data, &state); err != nil {
		log.Println(err)
		return state, errors.New("Unable to Translate tfstate in Memory")
	}
	return state, nil
}

// compares incoming resources from remote to what is already defined in the main.tf
// file to prevent duplicate definitions which breaks terraform import
func FilterExistingDefinitions(f io.Reader, importable tfimportables.Importable) ([]tfimportables.ResourceDefinition, []string) {
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

	uniqueResourceDefinitions := []tfimportables.ResourceDefinition{}
	uniqueProviders := []string{}
	providerMap := map[string]int{}
	resourceDefinitions := importable.ImportFromRemote()
	for _, resourceDefinition := range resourceDefinitions {
		providerMap[resourceDefinition.Provider]++
		if collection["resource"][fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)] == 0 {
			uniqueResourceDefinitions = append(uniqueResourceDefinitions, resourceDefinition)
		}
	}

	for provider := range providerMap {
		if collection["provider"][provider] == 0 {
			uniqueProviders = append(uniqueProviders, provider)
		}
	}

	return uniqueResourceDefinitions, uniqueProviders
}

// in preparation for terraform import, appends empty resource definitions to the existing main.tf file
func WriteHCLDefinitionHeaders(resourceDefinitions []tfimportables.ResourceDefinition, providerDefinitions []string, planFile io.Writer) error {
	var builder strings.Builder
	for _, newProvider := range providerDefinitions {
		builder.WriteString(fmt.Sprintf("provider %s {\n\talias = \"%s\"\n}\n\n", newProvider, newProvider))
	}
	for _, resourceDefinition := range resourceDefinitions {
		builder.WriteString(fmt.Sprintf("resource %s %s {}\n", resourceDefinition.Type, resourceDefinition.Name))
	}
	if _, err := planFile.Write([]byte(builder.String())); err != nil {
		return err
	}
	return nil
}

// takes the tfstate representations formats them as HCL and writes them to a bytes buffer
// so it can be flushed into main.tf
func ConvertTFStateToHCL(state State, importable tfimportables.Importable) []byte {
	var builder strings.Builder
	knownProviders := map[string]int{}

	log.Println("Assembling main.tf...")

	for _, resource := range state.Resources {
		providerDefinition := strings.Replace(resource.Provider, "provider.", "", 1)
		if knownProviders[providerDefinition] == 0 {
			knownProviders[providerDefinition]++
			builder.WriteString(fmt.Sprintf("provider %s {\n\talias = \"%s\"\n}\n\n", providerDefinition, providerDefinition))
		}
		for _, instance := range resource.Instances {
			builder.WriteString(fmt.Sprintf("resource %s %s {\n", resource.Type, resource.Name))
			builder.WriteString(fmt.Sprintf("\tprovider = %s\n", providerDefinition))
			b, _ := json.Marshal(instance.Data)
			hclShape := importable.HCLShape()
			json.Unmarshal(b, hclShape)
			convertToHCLLine(hclShape, 1, &builder)
			builder.WriteString("}\n\n")
		}
		builder.WriteString(string(resource.Content))
	}
	return []byte(builder.String())
}

func indent(level int) []byte {
	out := make([]byte, level)
	for i := 0; i < level; i++ {
		out[i] = byte('\t')
	}
	return out
}

// recursively converts a chunk of data from it's struct representation to its HCL representation
// and appends the "line" to a bytes buffer.
func convertToHCLLine(input interface{}, indentLevel int, builder *strings.Builder) {
	b, err := json.Marshal(input)
	if err != nil {
		log.Fatalln("unable to parse state to hcl")
	}
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	for k, v := range m {
		if v != nil {
			switch reflect.TypeOf(v).Kind() {
			case reflect.String:
				builder.WriteString(fmt.Sprintf("%s%s = %q\n", indent(indentLevel), utils.ToSnakeCase(k), v))
			case reflect.Int, reflect.Int32, reflect.Float32, reflect.Float64, reflect.Bool:
				builder.WriteString(fmt.Sprintf("%s%s = %v\n", indent(indentLevel), utils.ToSnakeCase(k), v))
			case reflect.Array, reflect.Slice:
				sl := v.([]interface{})
				if len(sl) > 0 {
					switch reflect.TypeOf(sl[0]).Kind() { // array of complex stuff
					case reflect.Array, reflect.Slice, reflect.Map:
						for j := 0; j < len(sl); j++ {
							builder.WriteString(strings.ToLower(fmt.Sprintf("\n%s%s {\n", indent(indentLevel), utils.ToSnakeCase(k))))
							convertToHCLLine(sl[j], indentLevel+1, builder)
							builder.WriteString(fmt.Sprintf("%s}\n", indent(indentLevel)))
						}
					default: // array of strings
						builder.WriteString(fmt.Sprintf("%s%s = [", indent(indentLevel), utils.ToSnakeCase(k)))
						for j := 0; j < len(sl); j++ {
							builder.WriteString(fmt.Sprintf("%q", sl[j]))
							if j < len(sl)-1 {
								builder.WriteString(",")
							}
						}
						builder.WriteString("]\n")
					}
				}
			case reflect.Map:
				if len(v.(map[string]interface{})) > 0 {
					builder.WriteString(strings.ToLower(fmt.Sprintf("\n%s%s = {\n", indent(indentLevel), utils.ToSnakeCase(k))))
					convertToHCLLine(v, indentLevel+1, builder)
					builder.WriteString(fmt.Sprintf("%s}\n", indent(indentLevel)))
				}
			default:
				fmt.Println("Unable to Determine Type", k, v)
			}
		}
	}
}
