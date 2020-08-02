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
