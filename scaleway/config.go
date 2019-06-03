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

	"github.com/hashicorp/terraform/terraform"

	"github.com/scaleway/scaleway-sdk-go/scw"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform/helper/logging"
	deprecatedSdk "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/logger"
	"github.com/scaleway/scaleway-sdk-go/scwconfig"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

// scwConfig contains the Scaleway configuration.
var scwConfig scwconfig.Config

func init() {
	// Init the Scaleway config.
	config, err := scwconfig.Load()
	if err != nil {
		log.Fatalf("error: cannot load configuration: %s", err)
	}
	scwConfig = config

	// Init the SDK logger.
	logger.SetLogger(sdkLogger{})
}

// Config is a configuration for a client.
type Config struct {
	AccessKey             string
	SecretKey             string
	DefaultOrganizationID string
	DefaultRegion         utils.Region
	DefaultZone           utils.Zone
}

// Meta contains SDK clients used by resources.
//
// This meta value is passed into all resources.
type Meta struct {
	// The Scaleway SDK client.
	scwClient *scw.Client

	// Deprecated: The deprecated Scaleway SDK (will be removed in `v2.0.0`).
	deprecatedClient *deprecatedSdk.API
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

	// Fetch known scaleway server types to support validation in r/server
	err = fetchServerAvailabilities(client)
	if err != nil {
		log.Printf("error: cannot fetch server availabilities: %s", err)
	}

	return meta, nil
}

var userAgent = fmt.Sprintf("terraform-provider/%s;terraform/%s", version, terraform.VersionString())

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

	if c.DefaultOrganizationID != "" {
		options = append(options, scw.WithDefaultOrganizationID(c.DefaultOrganizationID))
	}

	if c.DefaultRegion != "" {
		options = append(options, scw.WithDefaultRegion(c.DefaultRegion))
	}

	if c.DefaultZone != "" {
		options = append(options, scw.WithDefaultZone(c.DefaultZone))
	}

	// TODO: Use retryablehttp client here
	client, err := scw.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("error: cannot create SDK client: %s", err)
	}

	return client, err
}

// GetDeprecatedClient create a new deprecated client from a configuration.
func (c *Config) GetDeprecatedClient() (*deprecatedSdk.API, error) {
	options := func(sdkApi *deprecatedSdk.API) {
		sdkApi.Client = createRetryableHTTPClient()
	}

	// TODO: Replace by a parsing with error handling.
	region := ""
	if c.DefaultRegion == utils.RegionFrPar || c.DefaultZone == utils.ZoneFrPar1 {
		region = "par1"
	}
	if c.DefaultRegion == utils.RegionNlAms || c.DefaultZone == utils.ZoneNlAms1 {
		region = "ams1"
	}

	return deprecatedSdk.New(
		c.DefaultOrganizationID,
		c.SecretKey,
		region,
		options,
	)
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
		if resp == nil {
			return true, err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
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
