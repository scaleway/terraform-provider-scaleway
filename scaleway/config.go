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

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform/helper/logging"
	sdk "github.com/nicolai86/scaleway-sdk"
)

// Config contains scaleway configuration values
type Config struct {
	Organization string
	APIKey       string
	Region       string
}

// Meta contains SDK clients used by resources.
//
// This meta value is passed into all resources.
type Meta struct {
	// Deprecated: The deprecated Scaleway SDK (will be removed in `v2.0.0`).
	deprecatedClient *sdk.API
}

// Meta creates a meta instance from a client configuration.
func (c *Config) Meta() (*Meta, error) {
	meta := &Meta{}

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
		cl := retryablehttp.NewClient()

		cl.HTTPClient.Transport = logging.NewTransport("Scaleway", cl.HTTPClient.Transport)
		cl.RetryMax = 3
		cl.RetryWaitMax = 2 * time.Minute
		cl.Logger = log.New(os.Stderr, "", 0)
		cl.RetryWaitMin = time.Minute
		cl.CheckRetry = func(_ context.Context, resp *http.Response, err error) (bool, error) {
			if resp == nil {
				return true, err
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				return true, err
			}
			return retryablehttp.DefaultRetryPolicy(context.TODO(), resp, err)
		}

		sdkApi.Client = &client{cl}
	}

	return sdk.New(
		c.Organization,
		c.APIKey,
		c.Region,
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
