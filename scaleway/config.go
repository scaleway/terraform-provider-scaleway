package scaleway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform/helper/logging"
	sdk "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

// Meta contains config and SDK clients used by resources.
//
// This meta value is passed into all resources.
type Meta struct {
	AccessKey        string
	SecretKey        string
	DefaultProjectID string
	DefaultRegion    utils.Region
	DefaultZone      utils.Zone

	// scwClient is the Scaleway SDK client.
	scwClient *scw.Client

	// Deprecated: deprecatedClient is the deprecated Scaleway SDK (will be removed in `v2.0.0`).
	deprecatedClient *sdk.API
}

// bootstrapScwClient returns a new scw.Client from the configuration.
func (m *Meta) bootstrapScwClient() error {
	options := []scw.ClientOption{
		scw.WithHTTPClient(createRetryableHTTPClient(false)),
		scw.WithUserAgent(userAgent),
	}

	// The access key is not used for API authentications.
	if m.SecretKey != "" {
		options = append(options, scw.WithAuth(m.AccessKey, m.SecretKey))
	}

	if m.DefaultProjectID != "" {
		options = append(options, scw.WithDefaultProjectID(m.DefaultProjectID))
	}

	if m.DefaultRegion != "" {
		options = append(options, scw.WithDefaultRegion(m.DefaultRegion))
	}

	if m.DefaultZone != "" {
		options = append(options, scw.WithDefaultZone(m.DefaultZone))
	}

	client, err := scw.NewClient(options...)
	if err != nil {
		return fmt.Errorf("cannot create SDK client: %s", err)
	}

	m.scwClient = client
	return nil
}

// createRetryableHTTPClient creates a retryablehttp.Client.
func createRetryableHTTPClient(shouldLog bool) *client {
	c := retryablehttp.NewClient()

	if shouldLog {
		c.HTTPClient.Transport = logging.NewTransport("Scaleway", c.HTTPClient.Transport)
	}
	c.RetryMax = 3
	c.RetryWaitMax = 2 * time.Minute
	c.Logger = log.New(os.Stderr, "", 0)
	c.RetryWaitMin = time.Second * 2
	c.CheckRetry = func(_ context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil || resp.StatusCode == http.StatusTooManyRequests {
			return true, err
		}
		return retryablehttp.DefaultRetryPolicy(context.TODO(), resp, err)
	}

	return &client{c}
}

// client is a bridge between scw.httpClient interface and retryablehttp.Client
type client struct {
	*retryablehttp.Client
}

// Do wraps calling an HTTP method with retries.
func (c *client) Do(r *http.Request) (*http.Response, error) {
	var body io.ReadSeeker
	if r.Body != nil {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bs)
	}
	req, err := retryablehttp.NewRequest(r.Method, r.URL.String(), body)
	if err != nil {
		return nil, err
	}
	for key, val := range r.Header {
		req.Header.Set(key, val[0])
	}
	return c.Client.Do(req)
}

// bootstrapDeprecatedClient creates a new deprecated client from the configuration.
func (m *Meta) bootstrapDeprecatedClient() error {
	options := func(sdkApi *sdk.API) {
		sdkApi.Client = createRetryableHTTPClient(true)
	}

	region := string(m.DefaultRegion)
	if m.DefaultRegion == utils.RegionFrPar {
		region = "par1"
	}
	if m.DefaultRegion == utils.RegionNlAms {
		region = "ams1"
	}

	sdk, err := sdk.New(
		m.DefaultProjectID,
		m.SecretKey,
		region,
		options,
	)
	if err != nil {
		return fmt.Errorf("cannot create deprecated SDK client: %s", err)
	}

	m.deprecatedClient = sdk
	return nil
}

// deprecatedScalewayConfig is the structure of the deprecated Scaleway config file.
type deprecatedScalewayConfig struct {
	Organization string `json:"organization"`
	Token        string `json:"token"`
	Version      string `json:"version"`
}

// readDeprecatedScalewayConfig parses the deprecated Scaleway config file.
func readDeprecatedScalewayConfig(path string) (string, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	var data deprecatedScalewayConfig
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return "", "", err
	}
	return data.Token, data.Organization, nil
}
