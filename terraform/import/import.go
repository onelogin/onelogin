package tfimport

import (
	"bufio"
	"fmt"
	"github.com/onelogin/onelogin/terraform/importables"
	"io"
	"regexp"
	"strings"
)

// compares incoming resources from remote to what is already defined in the main.tf
// file to prevent duplicate definitions which breaks terraform import
func FilterExistingDefinitions(f io.Reader, resources []tfimportables.ResourceDefinition) ([]tfimportables.ResourceDefinition, []string) {
	resourceDefinitionsToImport := []tfimportables.ResourceDefinition{} // resource definitions not in HCL file that were included in incoming resources
	unspecifiedProviders := []string{}                                  // providers not already in HCL file from which to import new resources

	// resource definition headers in HCL file like resource onelogin_apps cool_app {}
	searchCriteria := map[string]*regexp.Regexp{
		"provider": regexp.MustCompile(`(\w*provider\w*)\s(([a-zA-Z\_]*))\s\{`),
		"resource": regexp.MustCompile(`(\w*resource\w*)\s([a-zA-Z\_\-]*)\s([a-zA-Z\_\-]*[0-9]*)\s?\{`),
	}

	// running tab of provider and resource definitions in HCL file
	definitionHeaderCounter := map[string]map[string]int{
		"provider": map[string]int{},
		"resource": map[string]int{},
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		for regexName, r := range searchCriteria {
			definitionHeaderLine := r.FindStringSubmatch(t)
			if len(definitionHeaderLine) > 0 {
				var definitionKey string
				if regexName == "provider" {
					definitionKey = fmt.Sprintf("%s", definitionHeaderLine[len(definitionHeaderLine)-1])
				}
				if regexName == "resource" {
					definitionKey = fmt.Sprintf("%s.%s", definitionHeaderLine[len(definitionHeaderLine)-2], definitionHeaderLine[len(definitionHeaderLine)-1])
				}
				definitionHeaderCounter[regexName][definitionKey]++
			}
		}
	}

	for _, resourceDefinition := range resources {
		if definitionHeaderCounter["provider"][resourceDefinition.Provider] == 0 {
			definitionHeaderCounter["provider"][resourceDefinition.Provider]++
			unspecifiedProviders = append(unspecifiedProviders, resourceDefinition.Provider)
		}
		if definitionHeaderCounter["resource"][fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)] == 0 {
			resourceDefinitionsToImport = append(resourceDefinitionsToImport, resourceDefinition)
		}
	}

	return resourceDefinitionsToImport, unspecifiedProviders
}

// WriteHCLDefinitionHeaders appends empty resource definitions to the existing main.tf file so terraform import will pick them up
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
