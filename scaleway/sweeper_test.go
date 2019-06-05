package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedDeprecatedClientForRegion returns a scaleway deprecated client needed for the sweeper
// functions for a given region {par1,ams1}
func sharedDeprecatedClientForRegion(r string) (*api.API, error) {
	if os.Getenv("SCALEWAY_ORGANIZATION") == "" {
		return nil, fmt.Errorf("empty SCALEWAY_ORGANIZATION")
	}

	if os.Getenv("SCALEWAY_TOKEN") == "" {
		return nil, fmt.Errorf("empty SCALEWAY_TOKEN")
	}

	region, err := utils.ParseRegion(r)
	if err != nil {
		return nil, err
	}

	conf := &Config{
		Organization: os.Getenv("SCALEWAY_ORGANIZATION"),
		APIKey:       os.Getenv("SCALEWAY_TOKEN"),
		Region:       region,
	}

	// configures a default client for the region, using the above env vars
	client, err := conf.GetDeprecatedClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %#v", err)
	}

	return client, nil
}
