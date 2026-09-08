package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
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

	// RetryOn404WaitTimeout bounds how long RetryOn404 keeps retrying a transient HTTP 404, typically
	// to wait for consistency resolution in Secret/KM APIs.
	RetryOn404WaitTimeout = 30 * time.Second

	// RetryOn404WaitTime is the delay between retries on HTTP 404.
	RetryOn404WaitTime = 2 * time.Second
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
	c.RetryMax = 5
	c.RetryWaitMax = 2 * time.Minute
	c.Logger = logging.L
	c.RetryWaitMin = time.Second * 2
	c.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil || resp.StatusCode == http.StatusTooManyRequests {
			return true, err
		}

		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}
	c.Backoff = rateLimitBackoff

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

// rateLimitBackoff computes how long to wait before the next retry.
//
// Scaleway's API Gateway (Envoy) returns 429 responses carrying no timing
// guidance whatsoever — only "x-envoy-ratelimited: true", with no Retry-After
// or X-RateLimit-* headers. So client-side backoff is the only lever we have.
// The jitter is the important part: Terraform issues many requests concurrently,
// so without it every throttled request would retry in lock-step and immediately
// re-trigger the limit (a thundering herd). Retry-After is still honored if
// present, since some individual product APIs may set it and it is the standard.
func rateLimitBackoff(minWait, maxWait time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil {
		// Retry-After (RFC 7231): honored if present, though Scaleway's gateway
		// does not currently send it on rate-limited responses.
		if wait, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			return wait
		}
	}

	// Exponential backoff (minWait * 2^attemptNum), capped at maxWait.
	backoff := float64(minWait) * math.Pow(2, float64(attemptNum))
	if backoff > float64(maxWait) || math.IsInf(backoff, 0) {
		backoff = float64(maxWait)
	}

	// Full jitter across [minWait, backoff].
	span := int64(backoff) - int64(minWait)
	if span <= 0 {
		return minWait
	}

	return minWait + time.Duration(rand.Int63n(span))
}

// parseRetryAfter parses an RFC 7231 Retry-After value: either a number of
// seconds or an HTTP-date. It returns ok=false when the header is absent or
// malformed so the caller can fall back to computed backoff.
func parseRetryAfter(v string) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}

	if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
		if secs < 0 {
			return 0, false
		}

		return time.Duration(secs) * time.Second, true
	}

	if t, err := http.ParseTime(v); err == nil {
		if until := time.Until(t); until > 0 {
			return until, true
		}

		return 0, true // date already passed: retry immediately
	}

	return 0, false
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

// Retries the specified function when it returns a 404 error.
func RetryOn404[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	wait := RetryOn404WaitTime
	if DefaultWaitRetryInterval != nil {
		wait = *DefaultWaitRetryInterval
	}

	deadline := time.Now().Add(RetryOn404WaitTimeout)

	for {
		result, err := f(ctx)
		if err == nil {
			return result, nil
		}

		if !httperrors.Is404(err) || time.Now().After(deadline) {
			return result, err
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(wait):
		}
	}
}
