package acctest

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/provider"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/stretchr/testify/require"
)

func PreCheck(_ *testing.T) {}

type TestTools struct {
	T                 *testing.T
	Meta              *meta.Meta
	ProviderFactories map[string]func() (*schema.Provider, error)
	Cleanup           func()
}

func NewTestTools(t *testing.T) *TestTools {
	t.Helper()

	ctx := t.Context()

	folder, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot detect working directory for testing")
	}

	// Create a http client with recording capabilities
	httpClient, cleanup, err := getHTTPRecoder(t, folder, *UpdateCassettes)
	require.NoError(t, err)

	// Create meta that will be passed in the provider config
	m, err := meta.NewMeta(ctx, &meta.Config{
		ProviderSchema:   nil,
		TerraformVersion: "terraform-tests",
		HTTPClient:       httpClient,
	})
	require.NoError(t, err)

	if !*UpdateCassettes {
		tmp := 0 * time.Second
		transport.DefaultWaitRetryInterval = &tmp
	}

	return &TestTools{
		T:    t,
		Meta: m,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"scaleway": func() (*schema.Provider, error) {
				return provider.Provider(&provider.Config{Meta: m})(), nil
			},
		},
		Cleanup: cleanup,
	}
}

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

// compareJSONBodies compare two given maps that represent json bodies
// returns true if both json are equivalent
func compareJSONBodies(expected, actual map[string]interface{}) bool {
	// Check for each key in actual requests
	// Compare its value to cassette content if marshal-able to string
	for key := range actual {
		expectedValue, exists := expected[key]
		if !exists {
			// Actual request may contain a field that does not exist in cassette
			// New fields can appear in requests with new api features
			// We do not want to generate new cassettes for each new features
			continue
		}

		if !compareJSONFields(expectedValue, actual[key]) {
			return false
		}
	}

	for key := range expected {
		_, exists := actual[key]
		if !exists && expected[key] != nil {
			// Fails match if cassettes contains a field not in actual requests
			// Fields should not disappear from requests unless a sdk breaking change
			// We ignore if field is nil in cassette as it could be an old deprecated and unused field
			return false
		}
	}

	return true
}

// IsTestResource returns true if given resource identifier is from terraform test
// identifier should be resource name but some resource don't have names
// return true if identifier match regex "tf[-_]test"
// common used prefixes are "tf_tests", "tf_test", "tf-tests", "tf-test"
func IsTestResource(identifier string) bool {
	return len(identifier) >= len("tf_test") &&
		strings.HasPrefix(identifier, "tf") &&
		(identifier[2] == '_' || identifier[2] == '-') &&
		identifier[3:7] == "test"
}
