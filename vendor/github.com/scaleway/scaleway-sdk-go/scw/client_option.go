package scw

import (
	"net/http"
	"net/url"

	"github.com/scaleway/scaleway-sdk-go/internal/auth"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
)

// ClientOption is a function which applies options to a settings object.
type ClientOption func(*settings)

// httpClient wraps the net/http Client Do method
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// WithHTTPClient client option allows passing a custom http.Client which will be used for all requests.
func WithHTTPClient(httpClient httpClient) ClientOption {
	return func(s *settings) {
		s.httpClient = httpClient
	}
}

// WithoutAuth client option sets the client token to an empty token.
func WithoutAuth() ClientOption {
	return func(s *settings) {
		s.token = auth.NewNoAuth()
	}
}

// WithAuth client option sets the client access key and secret key.
func WithAuth(accessKey, secretKey string) ClientOption {
	return func(s *settings) {
		s.token = auth.NewToken(accessKey, secretKey)
	}
}

// WithAPIURL client option overrides the API URL of the Scaleway API to the given URL.
func WithAPIURL(apiURL string) ClientOption {
	return func(s *settings) {
		s.apiURL = apiURL
	}
}

// WithInsecure client option enables insecure transport on the client.
func WithInsecure() ClientOption {
	return func(s *settings) {
		s.insecure = true
	}
}

// WithUserAgent client option append a user agent to the default user agent of the SDK.
func WithUserAgent(ua string) ClientOption {
	return func(s *settings) {
		if s.userAgent != "" && ua != "" {
			s.userAgent += " "
		}
		s.userAgent += ua
	}
}

// withDefaultUserAgent client option overrides the default user agent of the SDK.
func withDefaultUserAgent(ua string) ClientOption {
	return func(s *settings) {
		s.userAgent = ua
	}
}

// WithProfile client option configures a client from the given profile.
func WithProfile(p *Profile) ClientOption {
	return func(s *settings) {
		accessKey := ""
		if p.AccessKey != nil {
			accessKey = *p.AccessKey
		}

		if p.SecretKey != nil {
			s.token = auth.NewToken(accessKey, *p.SecretKey)
		}

		if p.APIURL != nil {
			s.apiURL = *p.APIURL
		}

		if p.Insecure != nil {
			s.insecure = *p.Insecure
		}

		if p.DefaultProjectID != nil {
			projectID := *p.DefaultProjectID
			s.defaultProjectID = &projectID
		}

		if p.DefaultRegion != nil {
			defaultRegion := Region(*p.DefaultRegion)
			s.defaultRegion = &defaultRegion
		}

		if p.DefaultZone != nil {
			defaultZone := Zone(*p.DefaultZone)
			s.defaultZone = &defaultZone
		}
	}
}

// WithProfile client option configures a client from the environment variables.
func WithEnv() ClientOption {
	return WithProfile(LoadEnvProfile())
}

// WithDefaultProjectID client option sets the client default project ID.
//
// It will be used as the default value of the project_id field in all requests made with this client.
func WithDefaultProjectID(projectID string) ClientOption {
	return func(s *settings) {
		s.defaultProjectID = &projectID
	}
}

// WithDefaultRegion client option sets the client default region.
//
// It will be used as the default value of the region field in all requests made with this client.
func WithDefaultRegion(region Region) ClientOption {
	return func(s *settings) {
		s.defaultRegion = &region
	}
}

// WithDefaultZone client option sets the client default zone.
//
// It will be used as the default value of the zone field in all requests made with this client.
func WithDefaultZone(zone Zone) ClientOption {
	return func(s *settings) {
		s.defaultZone = &zone
	}
}

// WithDefaultPageSize client option overrides the default page size of the SDK.
//
// It will be used as the default value of the page_size field in all requests made with this client.
func WithDefaultPageSize(pageSize int32) ClientOption {
	return func(s *settings) {
		s.defaultPageSize = &pageSize
	}
}

// settings hold the values of all client options
type settings struct {
	apiURL           string
	token            auth.Auth
	userAgent        string
	httpClient       httpClient
	insecure         bool
	defaultProjectID *string
	defaultRegion    *Region
	defaultZone      *Zone
	defaultPageSize  *int32
}

func newSettings() *settings {
	return &settings{}
}

func (s *settings) apply(opts []ClientOption) {
	for _, opt := range opts {
		opt(s)
	}
}

func (s *settings) validate() error {
	var err error
	if s.token == nil {
		return errors.New("no credential option provided")
	}

	_, err = url.Parse(s.apiURL)
	if err != nil {
		return errors.Wrap(err, "invalid url %s", s.apiURL)
	}

	// TODO: Check ProjectID format
	if s.defaultProjectID != nil && *s.defaultProjectID == "" {
		return errors.New("default project id cannot be empty")
	}

	// TODO: Check Region format
	if s.defaultRegion != nil && *s.defaultRegion == "" {
		return errors.New("default region cannot be empty")
	}

	// TODO: Check Zone format
	if s.defaultZone != nil && *s.defaultZone == "" {
		return errors.New("default zone cannot be empty")
	}

	if s.defaultPageSize != nil && *s.defaultPageSize <= 0 {
		return errors.New("default page size cannot be <= 0")
	}

	return nil
}
