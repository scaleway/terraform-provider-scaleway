package scaleway

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UpdateCassettes will update all cassettes of a given test
var UpdateCassettes = flag.Bool("cassettes", os.Getenv("TF_UPDATE_CASSETTES") == "true", "Record Cassettes")

// QueryMatcherIgnore contains the list of query value that should be ignored when matching requests with cassettes
var QueryMatcherIgnore = []string{
	"organization_id",
}

// BodyMatcherIgnore contains the list of json body keys that should be ignored when matching requests with cassettes
var BodyMatcherIgnore = []string{
	"organization_id",
	"project_id",
}

func testAccPreCheck(_ *testing.T) {}

// getTestFilePath returns a valid filename path based on the go test name and suffix. (Take care of non fs friendly char)
func getTestFilePath(t *testing.T, suffix string) string {
	t.Helper()
	specialChars := regexp.MustCompile(`[\\?%*:|"<>. ]`)

	// Replace nested tests separators.
	fileName := strings.ReplaceAll(t.Name(), "/", "-")

	fileName = strcase.ToBashArg(fileName)

	// Replace special characters.
	fileName = specialChars.ReplaceAllLiteralString(fileName, "") + suffix

	// Remove prefix to simplify
	fileName = strings.TrimPrefix(fileName, "test-acc-scaleway-")

	return filepath.Join(".", "testdata", fileName)
}

// cassetteMatcher is a custom matcher that will juste check equivalence of request bodies
func cassetteBodyMatcher(actual *http.Request, expected cassette.Request) bool {
	if actual.Body == nil || actual.ContentLength == 0 {
		if expected.Body == "" {
			return true // Body match if both are empty
		}
		return false
	}
	actualBody, err := actual.GetBody()
	if err != nil {
		panic(fmt.Errorf("cassette body matcher: failed to copy actual body: %w", err))
	}
	actualRawBody, err := io.ReadAll(actualBody)
	if err != nil {
		panic(fmt.Errorf("cassette body matcher: failed to read actual body: %w", err))
	}

	if string(actualRawBody) == expected.Body {
		// Try to match raw bodies if they are not JSON (ex: cloud-init config)
		return true
	}
	actualJson := make(map[string]interface{})
	expectedJson := make(map[string]interface{})

	err = json.Unmarshal(actualRawBody, &actualJson)
	if err != nil {
		panic(fmt.Errorf("cassette body matcher: failed to parse json body: %w", err))
	}

	err = json.Unmarshal([]byte(expected.Body), &expectedJson)
	if err != nil {
		panic(fmt.Errorf("cassette body matcher: failed to parse cassette json body: %w", err))
	}

	for _, key := range BodyMatcherIgnore {
		delete(actualJson, key)
		delete(expectedJson, key)
	}

	// Check for each key in actual requests
	// Compare its value to cassette content if marshal-able to string
	for key := range actualJson {
		expectedValue, exists := expectedJson[key]
		if !exists {
			// Actual request may contain a field that does not exist in cassette
			// New fields can appear in requests with new api features
			// We do not want to generate new cassettes for each new features
			continue
		}
		switch actualValue := actualJson[key].(type) {
		case fmt.Stringer:
			if actualValue.String() != expectedValue.(fmt.Stringer).String() {
				return false
			}
		}
	}
	for key := range expectedJson {
		_, exists := actualJson[key]
		if !exists {
			// Fails match if cassettes contains a field not in actual requests
			// Fields should not disappear from requests unless a sdk breaking change
			return false
		}
	}

	return true
}

// cassetteMatcher is a custom matcher that check equivalence of a played request against a recorded one
// It compares method, path and query but will remove unwanted values from query
func cassetteMatcher(actual *http.Request, expected cassette.Request) bool {
	expectedURL, _ := url.Parse(expected.URL)
	actualURL := actual.URL
	actualURLValues := actualURL.Query()
	expectedURLValues := expectedURL.Query()
	for _, query := range QueryMatcherIgnore {
		actualURLValues.Del(query)
		expectedURLValues.Del(query)
	}
	actualURL.RawQuery = actualURLValues.Encode()
	expectedURL.RawQuery = expectedURLValues.Encode()

	return actual.Method == expected.Method &&
		actual.URL.Path == expectedURL.Path &&
		actualURL.RawQuery == expectedURL.RawQuery &&
		cassetteBodyMatcher(actual, expected)
}

// getHTTPRecoder creates a new httpClient that records all HTTP requests in a cassette.
// This cassette is then replayed whenever tests are executed again. This means that once the
// requests are recorded in the cassette, no more real HTTP requests must be made to run the tests.
//
// It is important to add a `defer cleanup()` so the given cassette files are correctly
// closed and saved after the requests.
func getHTTPRecoder(t *testing.T, update bool) (client *http.Client, cleanup func(), err error) {
	t.Helper()
	recorderMode := recorder.ModeReplaying
	if update {
		recorderMode = recorder.ModeRecording
	}

	// Setup recorder and scw client
	r, err := recorder.NewAsMode(getTestFilePath(t, ".cassette"), recorderMode, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add custom matcher for requests and cassettes
	r.SetMatcher(cassetteMatcher)

	// Add a filter which removes Authorization headers from all requests:
	r.AddFilter(func(i *cassette.Interaction) error {
		i.Request.Headers = i.Request.Headers.Clone()
		delete(i.Request.Headers, "x-auth-token")
		delete(i.Request.Headers, "X-Auth-Token")
		delete(i.Request.Headers, "Authorization")
		return nil
	})

	return &http.Client{Transport: r}, func() {
		assert.NoError(t, r.Stop()) // Make sure recorder is stopped once done with it
	}, nil
}

type TestTools struct {
	T                 *testing.T
	Meta              *Meta
	ProviderFactories map[string]func() (*schema.Provider, error)
	Cleanup           func()
}

func NewTestTools(t *testing.T) *TestTools {
	t.Helper()
	ctx := context.Background()
	// Create a http client with recording capabilities
	httpClient, cleanup, err := getHTTPRecoder(t, *UpdateCassettes)
	require.NoError(t, err)

	// Create meta that will be passed in the provider config
	meta, err := buildMeta(ctx, &metaConfig{
		providerSchema:   nil,
		terraformVersion: "terraform-tests",
		httpClient:       httpClient,
	})
	require.NoError(t, err)

	if !*UpdateCassettes {
		tmp := 0 * time.Second
		DefaultWaitRetryInterval = &tmp
	}

	return &TestTools{
		T:    t,
		Meta: meta,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"scaleway": func() (*schema.Provider, error) {
				return Provider(&ProviderConfig{Meta: meta})(), nil
			},
		},
		Cleanup: cleanup,
	}
}

func SkipBetaTest(t *testing.T) {
	t.Helper()
	if !terraformBetaEnabled {
		t.Skip("Skip test as beta is not enabled")
	}
}

func TestAccScalewayProvider_SSHKeys(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayProvider_SSHKeys"
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEEYrzDOZmhItdKaDAEqJQ4ORS2GyBMtBozYsK5kiXXX opensource@scaleway.com"

	ctx := context.Background()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ProviderFactories: func() map[string]func() (*schema.Provider, error) {
			metaProd, err := buildMeta(ctx, &metaConfig{
				terraformVersion: "terraform-tests",
				httpClient:       tt.Meta.httpClient,
			})
			require.NoError(t, err)

			metaDev, err := buildMeta(ctx, &metaConfig{
				terraformVersion: "terraform-tests",
				httpClient:       tt.Meta.httpClient,
			})
			require.NoError(t, err)

			return map[string]func() (*schema.Provider, error){
				"prod": func() (*schema.Provider, error) {
					return Provider(&ProviderConfig{Meta: metaProd})(), nil
				},
				"dev": func() (*schema.Provider, error) {
					return Provider(&ProviderConfig{Meta: metaDev})(), nil
				},
			}
		}(),
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_ssh_key" "prod" {
						provider   = "prod" 
						name 	   = "%[1]s"
						public_key = "%[2]s"
					}

					resource "scaleway_account_ssh_key" "dev" {
						provider   = "dev" 
						name 	   = "%[1]s"
						public_key = "%[2]s"
					}
				`, SSHKeyName, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists(tt, "scaleway_account_ssh_key.prod"),
					testAccCheckScalewayAccountSSHKeyExists(tt, "scaleway_account_ssh_key.dev"),
				),
			},
		},
	})
}

func TestAccScalewayProvider_InstanceIPZones(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	ctx := context.Background()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ProviderFactories: func() map[string]func() (*schema.Provider, error) {
			metaProd, err := buildMeta(ctx, &metaConfig{
				terraformVersion: "terraform-tests",
				forceZone:        scw.ZoneFrPar2,
				httpClient:       tt.Meta.httpClient,
			})
			require.NoError(t, err)

			metaDev, err := buildMeta(ctx, &metaConfig{
				terraformVersion: "terraform-tests",
				forceZone:        scw.ZoneFrPar1,
				httpClient:       tt.Meta.httpClient,
			})
			require.NoError(t, err)

			return map[string]func() (*schema.Provider, error){
				"prod": func() (*schema.Provider, error) {
					return Provider(&ProviderConfig{Meta: metaProd})(), nil
				},
				"dev": func() (*schema.Provider, error) {
					return Provider(&ProviderConfig{Meta: metaDev})(), nil
				},
			}
		}(),
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_instance_ip dev {
					  provider = "dev"
					}
					
					resource scaleway_instance_ip prod {
					  provider = "prod"
					}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists(tt, "scaleway_instance_ip.prod"),
					testAccCheckScalewayInstanceIPExists(tt, "scaleway_instance_ip.dev"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.prod", "zone", "fr-par-2"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.dev", "zone", "fr-par-1"),
				),
			},
		},
	})
}
