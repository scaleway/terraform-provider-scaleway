package scw

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/scaleway/scaleway-sdk-go/internal/auth"
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

// WithConfig client option configure a client with Scaleway configuration.
func WithConfig(config Config) ClientOption {
	return func(s *settings) {
		// The access key is not used for API authentications.
		accessKey, _ := config.GetAccessKey()
		secretKey, secretKeyExist := config.GetSecretKey()
		if secretKeyExist {
			s.token = auth.NewToken(accessKey, secretKey)
		}

		apiURL, exist := config.GetAPIURL()
		if exist {
			s.apiURL = apiURL
		}

		insecure, exist := config.GetInsecure()
		if exist {
			s.insecure = insecure
		}

		defaultProjectID, exist := config.GetDefaultProjectID()
		if exist {
			s.defaultProjectID = &defaultProjectID
		}

		defaultRegion, exist := config.GetDefaultRegion()
		if exist {
			s.defaultRegion = &defaultRegion
		}

		defaultZone, exist := config.GetDefaultZone()
		if exist {
			s.defaultZone = &defaultZone
		}
	}
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
		return fmt.Errorf("no credential option provided")
	}

	_, err = url.Parse(s.apiURL)
	if err != nil {
		return fmt.Errorf("invalid url %s: %s", s.apiURL, err)
	}

	// TODO: Check ProjectID format
	if s.defaultProjectID != nil && *s.defaultProjectID == "" {
		return fmt.Errorf("default project id cannot be empty")
	}

	// TODO: Check Region format
	if s.defaultRegion != nil && *s.defaultRegion == "" {
		return fmt.Errorf("default region cannot be empty")
	}

	// TODO: Check Zone format
	if s.defaultZone != nil && *s.defaultZone == "" {
		return fmt.Errorf("default zone cannot be empty")
	}

	if s.defaultPageSize != nil && *s.defaultPageSize <= 0 {
		return fmt.Errorf("default page size cannot be <= 0")
	}

	return nil
}
