// Package env contains a list of environment variables used to modify the behaviour of the provider
package env

const (
	// RetryDelay is a duration string (parsed with time.ParseDuration) will change how long the SDK will wait between to attempts
	RetryDelay = "TF_RETRY_DELAY"
	// UpdateCassettes if set to "true" will trigger the cassettes to be recorded
	UpdateCassettes = "TF_UPDATE_CASSETTES"
	// TestDomain is the DNS domain used during our tests
	TestDomain = "TF_TEST_DOMAIN"
	// TestDomainZone is the DNS zone used during our tests
	TestDomainZone = "TF_TEST_DOMAIN_ZONE"
	// AppendUserAgent is appended to the user agent of the underlying SDK go
	AppendUserAgent = "TF_APPEND_USER_AGENT"
	// AccDomainRegistration if set to "true" will trigger acceptance test for domain registration
	AccDomainRegistration = "TF_ACC_DOMAIN_REGISTRATION"
)
