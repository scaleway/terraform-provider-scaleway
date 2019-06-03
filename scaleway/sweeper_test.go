package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	api "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedDeprecatedClientForRegion returns a scaleway deprecated client needed for the sweeper
// functions for a given region {par1,ams1}
func sharedDeprecatedClientForRegion(region string) (*api.API, error) {
	organizationId, exists := scwConfig.GetDefaultOrganizationID()
	if !exists {
		return nil, fmt.Errorf("a default organization ID must be set for sweeper tests")
	}

	accessKey, exists := scwConfig.GetAccessKey()
	if !exists {
		return nil, fmt.Errorf("an access key must be set for sweeper tests")
	}

	config := &Config{
		AccessKey:             accessKey,
		DefaultOrganizationID: organizationId,
		DefaultRegion:         utils.Region(region),
	}

	// configures a default client for the region, using the above env vars
	client, err := config.GetDeprecatedClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway deprecated client for sweeper tests: %#v", err)
	}

	return client, nil
}

// sharedClientForRegion returns a scaleway client needed for the sweeper
// functions for a given region {par1,ams1}
func sharedClientForRegion(region string) (*scw.Client, error) {
	_, exists := scwConfig.GetDefaultOrganizationID()
	if !exists {
		return nil, fmt.Errorf("a default organization ID must be set for sweeper tests")
	}

	_, exists = scwConfig.GetAccessKey()
	if !exists {
		return nil, fmt.Errorf("an access key must be set for sweeper tests")
	}

	config := &Config{}

	// configures a default client for the region, using the above env vars
	client, err := config.GetScwClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client for sweeper tests: %#v", err)
	}

	return client, nil
}
