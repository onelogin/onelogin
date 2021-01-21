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
	providerDefinitions := []string{}

	// resource definition headers in HCL file like resource onelogin_apps cool_app {}
	searchCriteria := map[string]*regexp.Regexp{
		"provider": regexp.MustCompile(`^\s*?source\s?=\s?"[a-zA-Z]+\/[a-zA-Z]+"?`),
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
					definitionKey = strings.ReplaceAll(strings.ReplaceAll(strings.Split(t, "=")[1], "\"", ""), " ", "")
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
		}
		if definitionHeaderCounter["resource"][fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)] == 0 {
			resourceDefinitionsToImport = append(resourceDefinitionsToImport, resourceDefinition)
		}
	}

	for k := range definitionHeaderCounter["provider"] {
		providerDefinitions = append(providerDefinitions, k)
	}
	return resourceDefinitionsToImport, providerDefinitions
}

// WriteHCLDefinitionHeaders appends empty resource definitions to the existing main.tf file so terraform import will pick them up
func AddNewProvidersAndResourceHCL(planFile io.Reader, newResourceDefinitions []tfimportables.ResourceDefinition, newProviderDefinitions []string) string {
	var builder strings.Builder
	re := regexp.MustCompile(`(\w*resource\w*)\s([a-zA-Z\_\-]*)\s([a-zA-Z\_\-]*[0-9]*)\s?\{`)

	builder.WriteString(fmt.Sprintf("terraform {\n\trequired_providers {\n"))
	for _, newProvider := range newProviderDefinitions {
		p := strings.Split(newProvider, "/")[0]
		builder.WriteString(fmt.Sprintf("\t\t%s = {\n\t\t\tsource = \"%s\"\n\t\t}\n", p, newProvider))
	}
	builder.WriteString(fmt.Sprintf("\t}\n}\n"))

	scanner := bufio.NewScanner(planFile)
	shouldRead := false
	for scanner.Scan() {
		t := scanner.Text()
		m := re.FindStringSubmatch(t)
		if len(m) != 0 {
			shouldRead = true
		}
		if shouldRead {
			builder.WriteString(fmt.Sprintf("%s\n", t))
		}
	}

	for _, resourceDefinition := range newResourceDefinitions {
		builder.WriteString(fmt.Sprintf("resource %s %s {}\n", resourceDefinition.Type, resourceDefinition.Name))
	}

	return builder.String()
}
