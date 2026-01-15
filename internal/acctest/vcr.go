package acctest

import (
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

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/strcase"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/env"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// UpdateCassettes will update all cassettes of a given test
var UpdateCassettes = flag.Bool("cassettes", os.Getenv(env.UpdateCassettes) == "true", "Record Cassettes")

// SensitiveFields is a map with keys listing fields that should be anonymized
// value will be set in place of its old value
var SensitiveFields = map[string]any{
	"secret_key": "00000000-0000-0000-0000-000000000000",
}

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
	// function related fields
	"mnq_project_id",
	"mnq_region",
	"mnq_nats_account_id",
	"mnq_nats_subject",
}

// removeKeyRecursive removes a key from a map and all its nested maps
func removeKeyRecursive(m map[string]any, key string) {
	delete(m, key)

	for _, v := range m {
		if v, ok := v.(map[string]any); ok {
			removeKeyRecursive(v, key)
		}
	}
}

func BuildCassetteName(testName string, pkgFolder string, suffix string) string {
	specialChars := regexp.MustCompile(`[\\?%*:|"<>. ]`)

	// Replace nested tests separators.
	fileName := strings.ReplaceAll(testName, "/", "-")

	fileName = strcase.ToBashArg(fileName)

	// Replace special characters.
	fileName = specialChars.ReplaceAllLiteralString(fileName, "") + suffix

	// Remove prefix to simplify
	fileName = strings.TrimPrefix(fileName, "test-acc-")

	return filepath.Join(pkgFolder, "testdata", fileName)
}

// getTestFilePath returns a valid filename path based on the go test name and suffix. (Take care of non fs friendly char)
func getTestFilePath(t *testing.T, pkgFolder string, suffix string) string {
	t.Helper()

	return BuildCassetteName(t.Name(), pkgFolder, suffix)
}

// cassetteMatcher is a custom matcher that will juste check equivalence of request bodies
func cassetteBodyMatcher(request *http.Request, cassette cassette.Request) bool {
	if request.Body == nil || request.ContentLength == 0 {
		if cassette.Body == "" {
			return true // Body match if both are empty
		}

		if _, isFile := request.Body.(*os.File); isFile {
			return true // Body match if request is sending a file, maybe do more check here
		}

		return false
	}

	r, err := request.GetBody()
	if err != nil {
		logging.L.Errorf("cassette body matcher: failed to copy request body: %v", err)

		return false
	}

	requestBody, err := io.ReadAll(r)
	if err != nil {
		logging.L.Errorf("cassette body matcher: failed to read actualRequest body: %v", err)

		return false
	}

	// Try to match raw bodies if they are not JSON (ex: cloud-init config)
	if string(requestBody) == cassette.Body {
		return true
	}

	requestJSON := make(map[string]any)
	cassetteJSON := make(map[string]any)

	// match if content is xml
	err = xml.Unmarshal(requestBody, new(any))
	if err == nil {
		return true
	}

	if !json.Valid(requestBody) {
		requestValues, err := url.ParseQuery(string(requestBody))
		if err != nil {
			logging.L.Errorf("cassette body matcher: failed to parse body as url values: %v", err)

			return false
		}

		// Remove keys that should be ignored during comparison
		for _, key := range BodyMatcherIgnore {
			requestValues.Del(key)
		}

		return compareFormBodies(requestValues, cassette.Form)
	}

	err = json.Unmarshal(requestBody, &requestJSON)
	if err != nil {
		logging.L.Errorf("cassette body matcher: failed to parse request body as json: %v", err)

		return false
	}

	err = json.Unmarshal([]byte(cassette.Body), &cassetteJSON)
	if err != nil {
		// actualRequest contains JSON but cassette may not contain JSON, this doesn't match in this case
		return false
	}
	// remove keys that should be ignored during comparison
	for _, key := range BodyMatcherIgnore {
		removeKeyRecursive(requestJSON, key)
		removeKeyRecursive(cassetteJSON, key)
	}

	return compareJSONBodies(requestJSON, cassetteJSON, false)
}

// CassetteMatcher is a custom matcher that check equivalence of a played request against a recorded one
// It compares method, path and query but will remove unwanted values from query
func CassetteMatcher(request *http.Request, cassette cassette.Request) bool {
	cassetteURL, _ := url.Parse(cassette.URL)
	requestURL := request.URL

	requestURLValues := requestURL.Query()
	cassetteURLValues := cassetteURL.Query()

	for _, query := range QueryMatcherIgnore {
		requestURLValues.Del(query)
		cassetteURLValues.Del(query)
	}

	requestURL.RawQuery = requestURLValues.Encode()
	cassetteURL.RawQuery = cassetteURLValues.Encode()

	// Specific handling of s3 URLs
	// Url format is https://test-acc-scaleway-object-bucket-lifecycle-8445817190507446251.s3.fr-par.scw.cloud/?lifecycle=
	if strings.HasSuffix(requestURL.Host, "scw.cloud") {
		if !strings.HasSuffix(cassetteURL.Host, "scw.cloud") {
			return false
		}

		requestS3Host := strings.Split(requestURL.Host, ".")
		cassetteS3Host := strings.Split(cassetteURL.Host, ".")

		if len(requestS3Host) >= 5 && len(cassetteS3Host) >= 5 {
			// Host is bucket.s3.region.scw.cloud
			// it could be a host without bucket name (ex: function upload)
			requestBucket := requestS3Host[0]
			cassetteBucket := cassetteS3Host[0]

			// Remove random number at the end of the bucket name
			if strings.Contains(requestBucket, "-") {
				requestBucket = requestBucket[:strings.LastIndex(requestBucket, "-")]
			}

			if strings.Contains(cassetteBucket, "-") {
				cassetteBucket = cassetteBucket[:strings.LastIndex(cassetteBucket, "-")]
			}

			if requestBucket != cassetteBucket {
				return false
			}
		}
	}

	return request.Method == cassette.Method &&
		request.URL.Path == cassetteURL.Path &&
		requestURL.RawQuery == cassetteURL.RawQuery &&
		cassetteBodyMatcher(request, cassette)
}

func cassetteSensitiveFieldsAnonymizer(i *cassette.Interaction) error {
	var jsonBody map[string]any

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
func getHTTPRecoder(t *testing.T, pkgFolder string, update bool) (client *http.Client, cleanup func(), err error) {
	t.Helper()

	recorderMode := recorder.ModeReplayOnly
	if update {
		recorderMode = recorder.ModeRecordOnly
	}

	cassetteFilePath := getTestFilePath(t, pkgFolder, ".cassette")
	_, errorCassette := os.Stat(cassetteFilePath + ".yaml")
	logging.L.Debugf("using %s.yaml", cassetteFilePath)

	// If in record mode we check that the cassette exists
	if recorderMode == recorder.ModeReplayOnly && errorCassette != nil {
		return nil, nil, fmt.Errorf("cannot stat file %s.yaml while in replay mode", cassetteFilePath)
	}

	// Setup recorder and scw client
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       getTestFilePath(t, pkgFolder, ".cassette"),
		Mode:               recorderMode,
		SkipRequestLatency: true,
	})
	if err != nil {
		return nil, nil, err
	}
	defer func(r *recorder.Recorder) {
		_ = r.Stop()
	}(r)

	// Add custom matcher for requests and cassettes
	r.SetMatcher(CassetteMatcher)

	// Add a filter which removes Authorization headers from all requests:
	r.AddHook(func(i *cassette.Interaction) error {
		i.Request.Headers = i.Request.Headers.Clone()
		delete(i.Request.Headers, "x-auth-token")
		delete(i.Request.Headers, "X-Auth-Token")
		delete(i.Request.Headers, "Authorization")

		return nil
	}, recorder.BeforeSaveHook)

	// Add a filter that will replace sensitive values with fixed values
	r.AddHook(cassetteSensitiveFieldsAnonymizer, recorder.BeforeSaveHook)

	retryOptions := transport.RetryableTransportOptions{}
	if !*UpdateCassettes {
		retryOptions.RetryWaitMax = scw.TimeDurationPtr(0)
	}

	return &http.Client{Transport: transport.NewRetryableTransportWithOptions(r, retryOptions)}, func() {
		require.NoError(t, r.Stop()) // Make sure recorder is stopped once done with it
	}, nil
}
