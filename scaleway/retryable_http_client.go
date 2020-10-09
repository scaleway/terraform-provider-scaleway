package scaleway

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

// TODO Retry logic should be moved in the SDK
// createRetryableHTTPClient creates a retryablehttp.Client.
func createRetryableHTTPClient(shouldLog bool) *client {
	c := retryablehttp.NewClient()

	if shouldLog {
		c.HTTPClient.Transport = logging.NewTransport("Scaleway", c.HTTPClient.Transport)
	}
	c.RetryMax = 3
	c.RetryWaitMax = 2 * time.Minute
	c.Logger = l
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
