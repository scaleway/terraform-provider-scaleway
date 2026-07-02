package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

const (
	// IAMPropagationTimeout bounds how long RetryOn403 keeps retrying a transient HTTP 403 while IAM
	// permissions propagate. IAM is eventually consistent and its cache is per-instance, so a freshly
	// granted permission can intermittently 403 even after a prior 200.
	IAMPropagationTimeout = 2 * time.Minute

	// RetryOn403WaitTime is the delay between retries on HTTP 403.
	RetryOn403WaitTime = 2 * time.Second
)

// DefaultWaitRetryInterval is used to set the retry interval to 0 during acceptance tests
var DefaultWaitRetryInterval *time.Duration

type RetryableTransportOptions struct {
	RetryMax     *int
	RetryWaitMax *time.Duration
	RetryWaitMin *time.Duration
}

func NewRetryableTransportWithOptions(defaultTransport http.RoundTripper, options RetryableTransportOptions) http.RoundTripper {
	c := retryablehttp.NewClient()
	c.HTTPClient = &http.Client{Transport: defaultTransport}

	// Defaults
	c.RetryMax = 3
	c.RetryWaitMax = 2 * time.Minute
	c.Logger = logging.L
	c.RetryWaitMin = time.Second * 2
	c.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil || resp.StatusCode == http.StatusTooManyRequests {
			return true, err
		}

		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	// If ErrorHandler is not set, retryablehttp will wrap http errors
	c.ErrorHandler = func(resp *http.Response, err error, _ int) (*http.Response, error) {
		// err is not nil if there was an error while performing request
		// it should be passed, but do not create an error when request contains an error code
		// http errors are handled by sdk coming after this transport
		if err != nil {
			return resp, err
		}

		return resp, nil
	}

	if options.RetryMax != nil {
		c.RetryMax = *options.RetryMax
	}

	if options.RetryWaitMax != nil {
		c.RetryWaitMax = *options.RetryWaitMax
	}

	if options.RetryWaitMin != nil {
		c.RetryWaitMin = *options.RetryWaitMin
	}

	return &RetryableTransport{c}
}

// NewRetryableTransport creates a http transport with retry capability.
// TODO Retry logic should be moved in the SDK
func NewRetryableTransport(defaultTransport http.RoundTripper) http.RoundTripper {
	return NewRetryableTransportWithOptions(defaultTransport, RetryableTransportOptions{})
}

// RetryableTransport client is a bridge between scw.httpClient interface and retryablehttp.Client
type RetryableTransport struct {
	*retryablehttp.Client
}

// RoundTrip wraps calling an HTTP method with retries.
func (c *RetryableTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var body io.ReadSeeker

	if r.Body != nil {
		bs, err := io.ReadAll(r.Body)
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

	req.GetBody = func() (io.ReadCloser, error) {
		b, err := req.BodyBytes()
		if err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(b)), err
	}

	return c.Do(req)
}

// RetryOn403 retries fn while it returns a transient HTTP 403 caused by IAM permission propagation,
// up to IAMPropagationTimeout. Any non-403 error (including a legitimate, persistent 403 that
// outlives the propagation window) is returned to the caller without further retries.
func RetryOn403(ctx context.Context, fn func() error) error {
	_, err := RetryOn403Value(ctx, func() (struct{}, error) {
		return struct{}{}, fn()
	})

	return err
}

// RetryOn403Value behaves like RetryOn403 for a function returning a value.
func RetryOn403Value[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	var (
		result  T
		lastErr error
	)

	wait := RetryOn403WaitTime
	if DefaultWaitRetryInterval != nil {
		wait = *DefaultWaitRetryInterval
	}

	deadline := time.Now().Add(IAMPropagationTimeout)

	for {
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		// Stop on any non-403 error, or once the propagation window elapses: a 403 outliving
		// IAMPropagationTimeout is treated as a real permission problem, not propagation lag.
		if !httperrors.Is403(lastErr) || time.Now().After(deadline) {
			return result, lastErr
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(wait):
		}
	}
}

func RetryOnTransientStateError[T any, U any](action func() (T, error), waiter func() (U, error)) (T, error) {
	t, err := action()

	if _, ok := errors.AsType[*scw.TransientStateError](err); ok {
		_, err := waiter()
		if err != nil {
			return t, err
		}

		return RetryOnTransientStateError(action, waiter)
	}

	return t, err
}
