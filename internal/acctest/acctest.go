package acctest

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/vcr"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/provider"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/stretchr/testify/require"
)

// UpdateCassettes will update all cassettes of a given test
var UpdateCassettes = flag.Bool("cassettes", os.Getenv("TF_UPDATE_CASSETTES") == "true", "Record Cassettes")

func PreCheck(_ *testing.T) {}

type TestTools struct {
	T                 *testing.T
	Meta              *meta.Meta
	ProviderFactories map[string]func() (*schema.Provider, error)
	Cleanup           func()
}

func NewRecordedClient(t *testing.T, pkgFolder string, update bool) (client *http.Client, cleanup func(), err error) {
	t.Helper()

	r, err := vcr.NewHTTPRecorder(t, pkgFolder, update)
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
	httpClient, cleanup, err := NewRecordedClient(t, folder, *UpdateCassettes)
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
