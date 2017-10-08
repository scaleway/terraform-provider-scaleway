package scaleway

import (
	"sort"

	"github.com/nicolai86/scaleway-sdk/api"
)

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

	// fetch known scaleway server types to support validation in r/server
	if len(commercialServerTypes) == 0 {
		if availability, err := api.GetServerAvailabilities(); err == nil {
			commercialServerTypes = availability.CommercialTypes()
			sort.StringSlice(commercialServerTypes).Sort()
		}
	}
	return &Client{api}, nil
}
