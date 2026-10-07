package models

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ResourceTemplate struct {
	LocalityAdjectiveUpper  string // Regional/Zoned
	LocalityAdjective       string // regional/zoned
	LocalityUpper           string // Region
	Locality                string // region
	Resource                string // FunctionNamespace
	ResourceClean           string // Namespace
	ResourceCleanLow        string // namespace
	ResourceFistLetterUpper string
	ResourceHCL             string // function_namespace
	API                     string // function
	APIFirstLetterUpper     string // Function

	SupportWaiters bool // If resource have waiters
}

func isUpper(letter uint8) bool {
	r := rune(letter)
	return unicode.IsUpper(r) || unicode.IsDigit(r)
}

func FirstLetterUpper(string string) string {
	capitalized := cases.Title(language.English).String(string)
	return capitalized
}

// splitByWord split a resource name to words
// ex: FunctionNamespace, K8SCluster, IamSSHKey
func splitByWord(sentence string) []string {
	words := []string(nil)
	wordBegin := 0
	prevLetterIsUpper := false
	for i := range sentence {
		nextLetterIsUpper := i == len(sentence)-1 || isUpper(sentence[i+1])
		currentLetterIsUpper := isUpper(sentence[i])

		// Try to detect first letter of a word
		if i != 0 && (currentLetterIsUpper && (!prevLetterIsUpper || !nextLetterIsUpper)) {
			words = append(words, sentence[wordBegin:i])
			wordBegin = i
		}

		// End of string
		if i == len(sentence)-1 {
			words = append(words, sentence[wordBegin:i+1])
			return words
		}

		prevLetterIsUpper = isUpper(sentence[i])
	}

	return words
}

func cleanResource(api string, resource string, upperCase bool) string {
	words := splitByWord(resource)

	if strings.ToLower(words[0]) == strings.ToLower(api) {
		words = words[1:]
	}

	if !upperCase {
		for i := range words {
			words[i] = strings.ToLower(words[i])
		}
	}

	return strings.Join(words, "")
}

func resourceWordsLower(resource string) []string {
	words := splitByWord(resource)

	for i := range words {
		words[i] = strings.ToLower(words[i])
	}

	return words
}

func adjectiveLocality(locality string) string {
	switch locality {
	case "zone":
		return "zonal"
	case "region":
		return "regional"
	}

	return locality
}

// api: function, container, instance
// resource: FunctionNamespace, InstanceServer, ContainerDomain
// locality: region, zone
func NewResourceTemplate(api string, resource string, locality string) ResourceTemplate {
	return ResourceTemplate{
		LocalityAdjectiveUpper: strings.Title(adjectiveLocality(locality)),
		LocalityAdjective:      adjectiveLocality(locality),
		LocalityUpper:          strings.Title(locality),
		Locality:               locality,
		Resource:               resource,
		ResourceClean:          cleanResource(api, resource, true),
		ResourceCleanLow:       cleanResource(api, resource, false),
		ResourceHCL:            strings.Join(resourceWordsLower(resource), "_"),
		API:                    api,
		APIFirstLetterUpper:    FirstLetterUpper(api),
	}
}
