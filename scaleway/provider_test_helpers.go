package scaleway

import (
	"strconv"
	"strings"
)

// Test Generated name has format: "{prefix}-{generated_number}
// example: test-acc-scaleway-project-3723338038624371236
func extractTestGeneratedNamePrefix(name string) string {
	// {prefix}-{generated}
	//         ^
	dashIndex := strings.LastIndex(name, "-")

	generated := name[dashIndex+1:]
	_, generatedToIntErr := strconv.ParseInt(generated, 10, 64)

	if dashIndex == -1 || generatedToIntErr != nil {
		// some are only {name}
		return name
	}

	// {prefix}
	return name[:dashIndex]
}

// Generated names have format: "tf-{prefix}-{generated1}-{generated2}"
// example: tf-sg-gifted-yonath
func extractGeneratedNamePrefix(name string) string {
	if strings.Count(name, "-") < 3 {
		return name
	}
	// tf-{prefix}-gifted-yonath
	name = strings.TrimPrefix(name, "tf-")

	// {prefix}-gifted-yonath
	//                ^
	dashIndex := strings.LastIndex(name, "-")
	name = name[:dashIndex]
	// {prefix}-gifted
	//         ^
	dashIndex = strings.LastIndex(name, "-")
	name = name[:dashIndex]
	return name
}

// compareJSONFieldsStrings compare two strings from request JSON bodies
// has special case when string are terraform generated names
func compareJSONFieldsStrings(expected, actual string) bool {
	expectedHandled := expected
	actualHandled := actual

	// Remove s3 url suffix to allow comparison
	if strings.HasSuffix(actual, ".s3-website.fr-par.scw.cloud") {
		actual = strings.TrimSuffix(actual, ".s3-website.fr-par.scw.cloud")
		expected = strings.TrimSuffix(expected, ".s3-website.fr-par.scw.cloud")
	}

	// Try to parse test generated name
	if strings.Contains(actual, "-") {
		expectedHandled = extractTestGeneratedNamePrefix(expected)
		actualHandled = extractTestGeneratedNamePrefix(actual)
	}

	// Try provider generated name
	if actualHandled == actual && strings.HasPrefix(actual, "tf-") {
		expectedHandled = extractGeneratedNamePrefix(expected)
		actualHandled = extractGeneratedNamePrefix(actual)
	}

	return expectedHandled == actualHandled
}
