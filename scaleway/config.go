package scaleway

import (
	"sort"

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

// Client configures and returns a fully initialized Scaleway client
func (c *Config) Client() (*Client, error) {
	api, err := sdk.New(
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
