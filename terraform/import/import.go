package tfimport

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	tfimportables "github.com/onelogin/onelogin/terraform/importables"
)

var providerRegex *regexp.Regexp = regexp.MustCompile(`^\s*?source\s?=\s?"[a-zA-Z]+\/[a-zA-Z]+"?`)
var resourceRegex *regexp.Regexp = regexp.MustCompile(`(\w*resource\w*)\s([a-zA-Z\_\-]*)\s([a-zA-Z0-9\_\-]*[0-9]*)\s?\{`)

func providerFromRegexMatches(matches []string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.Split(matches[0], "=")[1], "\"", ""), " ", "")
}

func resourceFromRegexMatches(matches []string) string {
	return fmt.Sprintf("%s.%s", matches[len(matches)-2], matches[len(matches)-1])
}

// DetermineNewResourcesAndProviders checks resource list from remote and cross references existing tf plan file to see which new resources and providers need to be added and imported
func DetermineNewResourcesAndProviders(planFile io.Reader, resourcesFromRemote []tfimportables.ResourceDefinition) ([]tfimportables.ResourceDefinition, []string) {
	// running tab of provider and resource definitions in HCL file
	definitionHeaderCounter := map[string]map[string]int{"provider": {}, "resource": {}}

	scanner := bufio.NewScanner(planFile)
	for scanner.Scan() {
		line := scanner.Text()
		providerLineMatches := providerRegex.FindStringSubmatch(line)
		resourceLineMatches := resourceRegex.FindStringSubmatch(line)
		if len(providerLineMatches) > 0 {
			definitionKey := providerFromRegexMatches(providerLineMatches)
			definitionHeaderCounter["provider"][definitionKey]++
		}
		if len(resourceLineMatches) > 0 {
			definitionKey := resourceFromRegexMatches(resourceLineMatches)
			definitionHeaderCounter["resource"][definitionKey]++
		}
	}

	// loop over all resources from remote and any that we dont see in the plan file should be tallied and added as a definition to import
	resourceDefinitionsToImport := []tfimportables.ResourceDefinition{}
	for _, resourceDefinition := range resourcesFromRemote {
		if definitionHeaderCounter["provider"][resourceDefinition.Provider] == 0 {
			definitionHeaderCounter["provider"][resourceDefinition.Provider]++
		}
		if definitionHeaderCounter["resource"][fmt.Sprintf("%s.%s", resourceDefinition.Type, resourceDefinition.Name)] == 0 {
			resourceDefinitionsToImport = append(resourceDefinitionsToImport, resourceDefinition)
		}
	}

	providerDefinitions := []string{}
	for providerName := range definitionHeaderCounter["provider"] {
		providerDefinitions = append(providerDefinitions, providerName)
	}
	return resourceDefinitionsToImport, providerDefinitions
}

// AddNewProvidersAndResourceHCL appends empty resource definitions to the existing main.tf file so terraform import will pick them up
func AddNewProvidersAndResourceHCL(planFile io.Reader, newResourceDefinitions []tfimportables.ResourceDefinition, newProviderDefinitions []string) string {
	var builder strings.Builder // builder represents the new state of our .tf file copied over from the existing tf file

	// we'll add the provider source headers for any new providers we dont know about
	builder.WriteString(fmt.Sprintf("terraform {\n\trequired_providers {\n"))
	for _, newProvider := range newProviderDefinitions {
		p := strings.Split(newProvider, "/")[1]
		builder.WriteString(fmt.Sprintf("\t\t%s = {\n\t\t\tsource = \"%s\"\n\t\t}\n", p, newProvider))
	}
	builder.WriteString(fmt.Sprintf("\t}\n}\n"))

	// then we scan the existing .tf file reading in all the existing lines to our new copy (the string builder)
	scanner := bufio.NewScanner(planFile)
	shouldRead := false
	for scanner.Scan() {
		t := scanner.Text()
		m := resourceRegex.FindStringSubmatch(t)
		if len(m) != 0 {
			shouldRead = true
		}
		if shouldRead {
			builder.WriteString(fmt.Sprintf("%s\n", t))
		}
	}

	// finally we add the resource xxx xxx {} lines for the resources we want to import
	for _, resourceDefinition := range newResourceDefinitions {
		builder.WriteString(fmt.Sprintf("resource %s %s {}\n", resourceDefinition.Type, resourceDefinition.Name))
	}

	// this is the representation of the new state of the .tf file with empty resource headers ready for import
	return builder.String()
}
