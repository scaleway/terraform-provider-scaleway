package models

import (
	"strings"
	"unicode"
)

type ResourceTemplate struct {
	LocalityAdjectiveUpper string // Regional
	LocalityUpper          string // Region
	Locality               string // region
	Resource               string // FunctionNamespace
	ResourceClean          string // Namespace
	ResourceCleanLow       string // namespace
	ResourceHCL            string // function_namespace
	API                    string // function
}

func isUpper(letter uint8) bool {
	r := rune(letter)
	return unicode.IsUpper(r) || unicode.IsDigit(r)
}

// splitByWord split a resource name to words
// ex: FunctionNamespace, K8SCluster, IamSSHKey
func splitByWord(sentence string) []string {
	words := []string(nil)
	wordBegin := 0
	prevLetterIsUpper := false
	for i := range sentence {
		nextLetterIsUpper := i == len(sentence)-1 || isUpper(sentence[i+1])
		_ = nextLetterIsUpper
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

// api: function, container, instance
// resource: FunctionNamespace, InstanceServer, ContainerDomain
// locality: region, zone
func NewResourceTemplate(api string, resource string, locality string) ResourceTemplate {
	return ResourceTemplate{
		LocalityAdjectiveUpper: strings.Title(locality) + "al",
		LocalityUpper:          strings.Title(locality),
		Locality:               locality,
		Resource:               resource,
		ResourceClean:          cleanResource(api, resource, true),
		ResourceCleanLow:       cleanResource(api, resource, false),
		ResourceHCL:            strings.Join(resourceWordsLower(resource), "_"),
		API:                    api,
	}
}
