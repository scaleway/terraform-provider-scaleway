package scaleway

import "github.com/nicolai86/scaleway-sdk/api"

// Config contains scaleway configuration values
type Config struct {
	Organization string
	APIKey       string
	Region       string
}

// Client contains scaleway api clients
type Client struct {
	scaleway *api.ScalewayAPI
}

// Client configures and returns a fully initialized Scaleway client
func (c *Config) Client() (*Client, error) {
	api, err := api.New(
		c.Organization,
		c.APIKey,
		c.Region,
	)
	if err != nil {
		return nil, err
	}
	return &Client{api}, nil
}
