package acctest

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/vcr"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/env"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

type TestTools struct {
	T                 *testing.T
	Meta              *meta.Meta
	ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
	Cleanup           func()
}

var foldersUsingVCRv4 = []string{
	"audittrail",
	"account",
	"container",
	"instance",
	"k8s",
	"marketplace",
}

func FolderUsesVCRv4(fullFolderPath string) bool {
	fullPathSplit := strings.Split(fullFolderPath, "/")

	folder := fullPathSplit[len(fullPathSplit)-1]
	for _, migratedFolder := range foldersUsingVCRv4 {
		if migratedFolder == folder {
			return true
		}
	}

	return false
}

// s3Encoder encodes binary payloads as base64 because serialization changed on go-vcr.v4
func s3Encoder(i *cassette.Interaction) error {
	if !strings.HasSuffix(i.Request.Host, "scw.cloud") {
		return nil
	}

	if i.Request.Body != "" && i.Request.Headers.Get("Content-Type") == "application/octet-stream" {
		requestBody := []byte(i.Request.Body)
		if !json.Valid(requestBody) {
			err := xml.Unmarshal(requestBody, new(any))
			if err != nil {
				i.Request.Body = base64.StdEncoding.EncodeToString(requestBody)
			}
		}
	}

	if i.Response.Body != "" && i.Response.Headers.Get("Content-Type") == "binary/octet-stream" {
		responseBody := []byte(i.Response.Body)
		if !json.Valid(responseBody) {
			err := xml.Unmarshal(responseBody, new(any))
			if err != nil {
				i.Response.Body = base64.StdEncoding.EncodeToString(responseBody)
			}
		}
	}

	return nil
}

func NewRecordedClient(t *testing.T, pkgFolder string, update bool) (client *http.Client, cleanup func(), err error) {
	t.Helper()

	s3EncoderHook := vcr.AdditionalHook{
		HookFunc: s3Encoder,
		Kind:     recorder.AfterCaptureHook,
	}

	r, err := vcr.NewHTTPRecorder(t, pkgFolder, update, nil, s3EncoderHook)
	if err != nil {
		return nil, nil, err
	}

	retryOptions := transport.RetryableTransportOptions{}
	if !update {
		retryOptions.RetryWaitMax = scw.TimeDurationPtr(0)
	}

	return &http.Client{
			Transport: transport.NewRetryableTransportWithOptions(r, retryOptions),
		}, func() {
			require.NoError(t, r.Stop()) // Make sure recorder is stopped once done with it
		}, nil
}

func NewTestTools(t *testing.T) *TestTools {
	t.Helper()

	ctx := t.Context()

	folder, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot detect working directory for testing")
	}

	// Create an HTTP client with recording capabilities
	var (
		httpClient *http.Client
		cleanup    func()
	)

	if FolderUsesVCRv4(folder) {
		httpClient, cleanup, err = NewRecordedClient(t, folder, *UpdateCassettes)
	} else {
		httpClient, cleanup, err = getHTTPRecoder(t, folder, *UpdateCassettes)
	}

	require.NoError(t, err)

	// Create meta that will be passed in the provider config
	m, err := meta.NewMeta(ctx, &meta.Config{
		ProviderSchema:   nil,
		TerraformVersion: "terraform-tests",
		HTTPClient:       httpClient,
	})
	require.NoError(t, err)

	if !*UpdateCassettes {
		// If no recording is happening, the delay to retry interactions should be 0
		tmp := 0 * time.Second
		transport.DefaultWaitRetryInterval = &tmp
	} else if os.Getenv(env.RetryDelay) != "" {
		// Overriding the delay interval is helpful to reduce the amount of requests performed while waiting for a resource to be available
		tmp, err := time.ParseDuration(os.Getenv(env.RetryDelay))
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("delay retry set to: %v", tmp)
		transport.DefaultWaitRetryInterval = &tmp
	}

	return &TestTools{
		T:    t,
		Meta: m,
		ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"scaleway": func() (tfprotov6.ProviderServer, error) {
				providers, errProvider := provider.NewProviderList(ctx, &provider.Config{Meta: m})
				if errProvider != nil {
					return nil, errProvider
				}

				muxServer, errMux := tf6muxserver.NewMuxServer(ctx, providers...)
				if errMux != nil {
					return nil, errMux
				}

				return muxServer.ProviderServer(), nil
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
