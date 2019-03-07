package scaleway

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	sdk "github.com/nicolai86/scaleway-sdk"
)

// Config contains scaleway configuration values
type Config struct {
	Organization string
	APIKey       string
	Region       string
}

// Client contains scaleway api clients
type Client struct {
	scaleway *sdk.API
}

// client is a bridge between sdk.HTTPClient interface and retryablehttp.Client
type client struct {
	*retryablehttp.Client
}

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

// Client configures and returns a fully initialized Scaleway client
func (c *Config) Client() (*Client, error) {
	api, err := sdk.New(
		c.Organization,
		c.APIKey,
		c.Region,
		func(c *sdk.API) {
			cl := retryablehttp.NewClient()
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
			c.Client = &client{cl}
		},
	)
	if err != nil {
		return nil, err
	}

	// fetch known scaleway server types to support validation in r/server
	if len(commercialServerTypes) == 0 {
		if availability, err := api.GetServerAvailabilities(); err == nil {
			commercialServerTypes = availability.CommercialTypes()
			sort.StringSlice(commercialServerTypes).Sort()
		}
		if os.Getenv("DISABLE_SCALEWAY_SERVER_TYPE_VALIDATION") != "" {
			commercialServerTypes = commercialServerTypes[:0]
		}
	}
	return &Client{api}, nil
}
