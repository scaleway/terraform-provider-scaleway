package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedClientForRegion returns a common scaleway client needed for the sweeper
// functions for a given region {par1,ams1}
func sharedClientForRegion(region string) (interface{}, error) {
	if os.Getenv("SCALEWAY_ORGANIZATION") == "" {
		return nil, fmt.Errorf("empty SCALEWAY_ORGANIZATION")
	}

	if os.Getenv("SCALEWAY_TOKEN") == "" {
		return nil, fmt.Errorf("empty SCALEWAY_TOKEN")
	}

	conf := &Config{
		Organization: os.Getenv("SCALEWAY_ORGANIZATION"),
		APIKey:       os.Getenv("SCALEWAY_TOKEN"),
		Region:       region,
	}

	// configures a default client for the region, using the above env vars
	client, err := conf.Client()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %#v", err)
	}

	return client, nil
}
