package scaleway

import (
	"context"
	"encoding/json"
	"encoding/xml"
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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/strcase"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
	"organization", // like organization_id but deprecated
	"organization_id",
	"project_id",
	"project", // like project_id but should be deprecated
}

// SensitiveFields is a map with keys listing fields that should be anonymized
// value will be set in place of its old value
var SensitiveFields = map[string]interface{}{
	"secret_key": "00000000-0000-0000-0000-000000000000",
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

func compareJSONFields(expected, actualI interface{}) bool {
	switch actual := actualI.(type) {
	case string:
		if _, isString := expected.(string); !isString {
			return false
		}
		return compareJSONFieldsStrings(expected.(string), actual)
	default:
		// Consider equality when not handled
		return true
	}
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

// compareFormBodies compare two given url.Values
// returns true if both url.Values are equivalent
func compareFormBodies(expected, actual url.Values) bool {
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

// cassetteMatcher is a custom matcher that will juste check equivalence of request bodies
func cassetteBodyMatcher(actualRequest *http.Request, cassetteRequest cassette.Request) bool {
	if actualRequest.Body == nil || actualRequest.ContentLength == 0 {
		if cassetteRequest.Body == "" {
			return true // Body match if both are empty
		} else if _, isFile := actualRequest.Body.(*os.File); isFile {
			return true // Body match if request is sending a file, maybe do more check here
		}
		return false
	}

	actualBody, err := actualRequest.GetBody()
	if err != nil {
		panic(fmt.Errorf("cassette body matcher: failed to copy actualRequest body: %w", err)) // lintignore: R009
	}
	actualRawBody, err := io.ReadAll(actualBody)
	if err != nil {
		panic(fmt.Errorf("cassette body matcher: failed to read actualRequest body: %w", err)) // lintignore: R009
	}

	// Try to match raw bodies if they are not JSON (ex: cloud-init config)
	if string(actualRawBody) == cassetteRequest.Body {
		return true
	}

	actualJSON := make(map[string]interface{})
	cassetteJSON := make(map[string]interface{})

	err = xml.Unmarshal(actualRawBody, new(interface{}))
	if err == nil {
		// match if content is xml
		return true
	}

	if !json.Valid(actualRawBody) {
		values, err := url.ParseQuery(string(actualRawBody))
		if err != nil {
			panic(fmt.Errorf("cassette body matcher: failed to parse body as url values: %w", err)) // lintignore: R009
		}

		// Remove keys that should be ignored during compare
		for _, key := range BodyMatcherIgnore {
			values.Del(key)
		}

		// Compare url values
		return compareFormBodies(values, cassetteRequest.Form)
	}

	err = json.Unmarshal(actualRawBody, &actualJSON)
	if err != nil {
		panic(fmt.Errorf("cassette body matcher: failed to parse json body: %w", err)) // lintignore: R009
	}

	err = json.Unmarshal([]byte(cassetteRequest.Body), &cassetteJSON)
	if err != nil {
		// actualRequest contains JSON but cassette may not contain JSON, this doesn't match in this case
		return false
	}

	// Remove keys that should be ignored during compare
	for _, key := range BodyMatcherIgnore {
		delete(actualJSON, key)
		delete(cassetteJSON, key)
	}

	return compareJSONBodies(cassetteJSON, actualJSON)
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

	// Specific handling of s3 URLs
	// Url format is https://test-acc-scaleway-object-bucket-lifecycle-8445817190507446251.s3.fr-par.scw.cloud/?lifecycle=
	if strings.HasSuffix(actualURL.Host, "scw.cloud") {
		if !strings.HasSuffix(expectedURL.Host, "scw.cloud") {
			return false
		}
		actualS3Host := strings.Split(actualURL.Host, ".")
		expectedS3Host := strings.Split(expectedURL.Host, ".")

		if len(actualS3Host) >= 5 && len(expectedS3Host) >= 5 {
			// Host is bucket.s3.region.scw.cloud
			// it could be a host without bucket name (ex: function upload)
			actualBucket := actualS3Host[0]
			expectedBucket := expectedS3Host[0]

			// Remove random number at the end of the bucket name
			if strings.Contains(actualBucket, "-") {
				actualBucket = actualBucket[:strings.LastIndex(actualBucket, "-")]
			}
			if strings.Contains(expectedBucket, "-") {
				expectedBucket = expectedBucket[:strings.LastIndex(expectedBucket, "-")]
			}

			if actualBucket != expectedBucket {
				return false
			}
		}
	}

	return actual.Method == expected.Method &&
		actual.URL.Path == expectedURL.Path &&
		actualURL.RawQuery == expectedURL.RawQuery &&
		cassetteBodyMatcher(actual, expected)
}

func cassetteSensitiveFieldsAnonymizer(i *cassette.Interaction) error {
	var jsonBody map[string]interface{}
	err := json.Unmarshal([]byte(i.Response.Body), &jsonBody)
	if err != nil {
		//nolint:nilerr
		return nil
	}
	for key, value := range SensitiveFields {
		if _, ok := jsonBody[key]; ok {
			jsonBody[key] = value
		}
	}
	anonymizedBody, err := json.Marshal(jsonBody)
	if err != nil {
		return fmt.Errorf("failed to marshal anonymized body: %w", err)
	}
	i.Response.Body = string(anonymizedBody)
	return nil
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

	// Add a filter that will replace sensitive values with fixed values
	r.AddSaveFilter(cassetteSensitiveFieldsAnonymizer)

	retryOptions := transport.RetryableTransportOptions{}
	if !*UpdateCassettes {
		retryOptions.RetryWaitMax = scw.TimeDurationPtr(0)
	}

	return &http.Client{Transport: transport.NewRetryableTransportWithOptions(r, retryOptions)}, func() {
		require.NoError(t, r.Stop()) // Make sure recorder is stopped once done with it
	}, nil
}

type FakeSideProjectTerminateFunc func() error

// createFakeSideProject creates a temporary project with a temporary IAM application and policy.
//
// The returned function is a cleanup function that should be called when to delete the project.
func createFakeSideProject(tt *TestTools) (*accountV3.Project, *iam.APIKey, FakeSideProjectTerminateFunc, error) {
	terminateFunctions := []FakeSideProjectTerminateFunc{}
	terminate := func() error {
		for i := len(terminateFunctions) - 1; i >= 0; i-- {
			err := terminateFunctions[i]()
			if err != nil {
				return err
			}
		}

		return nil
	}

	projectName := sdkacctest.RandomWithPrefix("test-acc-scaleway-project")
	iamApplicationName := sdkacctest.RandomWithPrefix("test-acc-scaleway-iam-app")
	iamPolicyName := sdkacctest.RandomWithPrefix("test-acc-scaleway-iam-policy")

	projectAPI := accountV3.NewProjectAPI(tt.Meta.ScwClient())
	project, err := projectAPI.CreateProject(&accountV3.ProjectAPICreateProjectRequest{
		Name: projectName,
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return projectAPI.DeleteProject(&accountV3.ProjectAPIDeleteProjectRequest{
			ProjectID: project.ID,
		})
	})

	iamAPI := iam.NewAPI(tt.Meta.ScwClient())
	iamApplication, err := iamAPI.CreateApplication(&iam.CreateApplicationRequest{
		Name: iamApplicationName,
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return iamAPI.DeleteApplication(&iam.DeleteApplicationRequest{
			ApplicationID: iamApplication.ID,
		})
	})

	iamPolicy, err := iamAPI.CreatePolicy(&iam.CreatePolicyRequest{
		Name:          iamPolicyName,
		ApplicationID: types.ExpandStringPtr(iamApplication.ID),
		Rules: []*iam.RuleSpecs{
			{
				ProjectIDs:         &[]string{project.ID},
				PermissionSetNames: &[]string{"ObjectStorageReadOnly", "ObjectStorageObjectsRead", "ObjectStorageBucketsRead"},
			},
		},
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return iamAPI.DeletePolicy(&iam.DeletePolicyRequest{
			PolicyID: iamPolicy.ID,
		})
	})

	iamAPIKey, err := iamAPI.CreateAPIKey(&iam.CreateAPIKeyRequest{
		ApplicationID:    types.ExpandStringPtr(iamApplication.ID),
		DefaultProjectID: &project.ID,
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return iamAPI.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
			AccessKey: iamAPIKey.AccessKey,
		})
	})

	return project, iamAPIKey, terminate, nil
}

// createFakeIAMManager creates a temporary project with a temporary IAM application and policy manager.
//
// The returned function is a cleanup function that should be called when to delete the project.
func createFakeIAMManager(tt *TestTools) (*accountV3.Project, *iam.APIKey, FakeSideProjectTerminateFunc, error) {
	terminateFunctions := []FakeSideProjectTerminateFunc{}
	terminate := func() error {
		for i := len(terminateFunctions) - 1; i >= 0; i-- {
			err := terminateFunctions[i]()
			if err != nil {
				return err
			}
		}

		return nil
	}

	projectName := sdkacctest.RandomWithPrefix("test-acc-scaleway-project")
	iamApplicationName := sdkacctest.RandomWithPrefix("test-acc-scaleway-iam-app")
	iamPolicyName := sdkacctest.RandomWithPrefix("test-acc-scaleway-iam-policy")

	projectAPI := accountV3.NewProjectAPI(tt.Meta.ScwClient())
	project, err := projectAPI.CreateProject(&accountV3.ProjectAPICreateProjectRequest{
		Name: projectName,
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return projectAPI.DeleteProject(&accountV3.ProjectAPIDeleteProjectRequest{
			ProjectID: project.ID,
		})
	})

	iamAPI := iam.NewAPI(tt.Meta.ScwClient())
	iamApplication, err := iamAPI.CreateApplication(&iam.CreateApplicationRequest{
		Name: iamApplicationName,
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return iamAPI.DeleteApplication(&iam.DeleteApplicationRequest{
			ApplicationID: iamApplication.ID,
		})
	})

	iamPolicy, err := iamAPI.CreatePolicy(&iam.CreatePolicyRequest{
		Name:          iamPolicyName,
		ApplicationID: types.ExpandStringPtr(iamApplication.ID),
		Rules: []*iam.RuleSpecs{
			{
				OrganizationID:     &project.OrganizationID,
				PermissionSetNames: &[]string{"IAMManager"},
			},
		},
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return iamAPI.DeletePolicy(&iam.DeletePolicyRequest{
			PolicyID: iamPolicy.ID,
		})
	})

	iamAPIKey, err := iamAPI.CreateAPIKey(&iam.CreateAPIKeyRequest{
		ApplicationID:    types.ExpandStringPtr(iamApplication.ID),
		DefaultProjectID: &project.ID,
	})
	if err != nil {
		if err := terminate(); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, nil, err
	}
	terminateFunctions = append(terminateFunctions, func() error {
		return iamAPI.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
			AccessKey: iamAPIKey.AccessKey,
		})
	})

	return project, iamAPIKey, terminate, nil
}

// fakeSideProjectProviders creates a new provider alias "side" with a new Config that will use the
// given project and API key as default profile configuration.
//
// This is useful to test resources that need to create resources in another project.
func fakeSideProjectProviders(ctx context.Context, tt *TestTools, project *accountV3.Project, iamAPIKey *iam.APIKey) map[string]func() (*schema.Provider, error) {
	t := tt.T

	metaSide, err := meta.NewMeta(ctx, &meta.Config{
		TerraformVersion:    "terraform-tests",
		HTTPClient:          tt.Meta.HTTPClient(),
		ForceProjectID:      project.ID,
		ForceOrganizationID: project.OrganizationID,
		ForceAccessKey:      iamAPIKey.AccessKey,
		ForceSecretKey:      *iamAPIKey.SecretKey,
	})
	require.NoError(t, err)

	providers := map[string]func() (*schema.Provider, error){
		"side": func() (*schema.Provider, error) {
			return Provider(&ProviderConfig{Meta: metaSide})(), nil
		},
	}

	for k, v := range tt.ProviderFactories {
		providers[k] = v
	}

	return providers
}

type TestTools struct {
	T                 *testing.T
	Meta              *meta.Meta
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
	meta, err := meta.NewMeta(ctx, &meta.Config{
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
		Meta: meta,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"scaleway": func() (*schema.Provider, error) {
				return Provider(&ProviderConfig{Meta: meta})(), nil
			},
		},
		Cleanup: cleanup,
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
			metaProd, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			metaDev, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				HTTPClient:       tt.Meta.HTTPClient(),
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
		CheckDestroy: testAccCheckScalewayIamSSHKeyDestroy(tt),
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
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_account_ssh_key.prod"),
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_account_ssh_key.dev"),
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
			metaProd, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				ForceZone:        scw.ZoneFrPar2,
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			metaDev, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				ForceZone:        scw.ZoneFrPar1,
				HTTPClient:       tt.Meta.HTTPClient(),
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
		CheckDestroy: testAccCheckScalewayIamSSHKeyDestroy(tt),
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
