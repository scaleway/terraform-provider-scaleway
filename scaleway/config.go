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
	"sort"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform/helper/logging"
	sdk "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

// Config is a configuration for a client.
type Config struct {
	Organization string
	APIKey       string
	Region       utils.Region
	Zone         utils.Zone
}

// Meta contains SDK clients used by resources.
//
// This meta value is passed into all resources.
type Meta struct {
	// scwClient is the Scaleway SDK client.
	scwClient *scw.Client

	// Deprecated: deprecatedClient is the deprecated Scaleway SDK (will be removed in `v2.0.0`).
	deprecatedClient *sdk.API
}

// Meta creates a meta instance from a client configuration.
func (c *Config) Meta() (*Meta, error) {
	meta := &Meta{}

	// Scaleway Client
	client, err := c.GetScwClient()
	if err != nil {
		return nil, err
	}
	meta.scwClient = client

	// Deprecated Scaleway Client
	deprecatedClient, err := c.GetDeprecatedClient()
	if err != nil {
		return nil, fmt.Errorf("error: cannot create deprecated client: %s", err)
	}
	meta.deprecatedClient = deprecatedClient

	// fetch known scaleway server types to support validation in r/server
	if len(commercialServerTypes) == 0 {
		if availability, err := deprecatedClient.GetServerAvailabilities(); err == nil {
			commercialServerTypes = availability.CommercialTypes()
			sort.StringSlice(commercialServerTypes).Sort()
		}
		if os.Getenv("DISABLE_SCALEWAY_SERVER_TYPE_VALIDATION") != "" {
			commercialServerTypes = commercialServerTypes[:0]
		}
	}

	return meta, nil
}

// GetScwClient returns a new scw.Client from a configuration.
func (c *Config) GetScwClient() (*scw.Client, error) {
	options := []scw.ClientOption{
		scw.WithHTTPClient(createsRetryableHTTPClient()),
		scw.WithUserAgent(UserAgent),
	}

	// The access key is not used for API authentications.
	if c.APIKey != "" {
		options = append(options, scw.WithAuth("", c.APIKey))
	}

	if c.Organization != "" {
		options = append(options, scw.WithDefaultProjectID(c.Organization))
	}

	if c.Region != "" {
		options = append(options, scw.WithDefaultRegion(c.Region))
	}

	if c.Zone != "" {
		options = append(options, scw.WithDefaultZone(c.Zone))
	}

	client, err := scw.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("cannot create SDK client: %s", err)
	}

	return client, err
}

// createsRetryableHTTPClient create a retryablehttp.Client.
func createsRetryableHTTPClient() *client {
	c := retryablehttp.NewClient()

	c.HTTPClient.Transport = logging.NewTransport("Scaleway", c.HTTPClient.Transport)
	c.RetryMax = 3
	c.RetryWaitMax = 2 * time.Minute
	c.Logger = log.New(os.Stderr, "", 0)
	c.RetryWaitMin = time.Minute
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
	for key, val := range r.Header {
		req.Header.Set(key, val[0])
	}
	if err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// GetDeprecatedClient create a new deprecated client from a configuration.
func (c *Config) GetDeprecatedClient() (*sdk.API, error) {
	options := func(sdkApi *sdk.API) {
		sdkApi.Client = createsRetryableHTTPClient()
	}

	// TODO: Replace by a parsing with error handling.
	region := ""
	if c.Region == utils.RegionFrPar || c.Zone == utils.ZoneFrPar1 {
		region = "par1"
	}
	if c.Region == utils.RegionNlAms || c.Zone == utils.ZoneNlAms1 {
		region = "ams1"
	}

	return sdk.New(
		c.Organization,
		c.APIKey,
		region,
		options,
	)
}

// deprecatedScalewayConfig is the structure of the deprecated Scaleway config file.
type deprecatedScalewayConfig struct {
	Organization string `json:"organization"`
	Token        string `json:"token"`
	Version      string `json:"version"`
}

// readDeprecatedScalewayConfig parse the deprecated Scaleway config file.
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
