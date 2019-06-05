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
	AccessKey        string
	SecretKey        string
	DefaultProjectID string
	DefaultRegion    utils.Region
	DefaultZone      utils.Zone
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
		scw.WithHTTPClient(createRetryableHTTPClient()),
		scw.WithUserAgent(userAgent),
	}

	// The access key is not used for API authentications.
	if c.SecretKey != "" {
		options = append(options, scw.WithAuth(c.AccessKey, c.SecretKey))
	}

	if c.DefaultProjectID != "" {
		options = append(options, scw.WithDefaultProjectID(c.DefaultProjectID))
	}

	if c.DefaultRegion != "" {
		options = append(options, scw.WithDefaultRegion(c.DefaultRegion))
	}

	if c.DefaultZone != "" {
		options = append(options, scw.WithDefaultZone(c.DefaultZone))
	}

	client, err := scw.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("cannot create SDK client: %s", err)
	}

	return client, err
}

// createRetryableHTTPClient create a retryablehttp.Client.
func createRetryableHTTPClient() *client {
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
		sdkApi.Client = createRetryableHTTPClient()
	}

	region := string(c.DefaultRegion)
	if c.DefaultRegion == utils.RegionFrPar {
		region = "par1"
	}
	if c.DefaultRegion == utils.RegionNlAms {
		region = "ams1"
	}

	return sdk.New(
		c.DefaultProjectID,
		c.SecretKey,
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
