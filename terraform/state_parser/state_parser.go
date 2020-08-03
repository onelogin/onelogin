package stateparser

import (
	"encoding/json"
	"fmt"
	"github.com/onelogin/onelogin-go-sdk/pkg/utils"
	"github.com/onelogin/onelogin/terraform/importables"
	"log"
	"reflect"
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

// takes the tfstate representations formats them as HCL and writes them to a bytes buffer
// so it can be flushed into main.tf
func ConvertTFStateToHCL(state State, importable tfimportables.Importable, outHCLShapeOption string) []byte {
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
			hclShape := importable.HCLShape(outHCLShapeOption)
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
