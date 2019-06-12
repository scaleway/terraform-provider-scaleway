package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	api "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/utils"
	_ "github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedDeprecatedClientForRegion returns a scaleway deprecated client needed for the sweeper
// functions for a given region {par1,ams1}
func sharedDeprecatedClientForRegion(region string) (*api.API, error) {
	projectId := os.Getenv("SCW_DEFAULT_PROJECT_ID")
	if projectId == "" {
		projectId = os.Getenv("SCALEWAY_ORGANIZATION")
	}
	if projectId == "" {
		return nil, fmt.Errorf("empty SCW_DEFAULT_PROJECT_ID")
	}

	secretKey := os.Getenv("SCW_SECRET_KEY")
	if secretKey == "" {
		secretKey = os.Getenv("SCALEWAY_TOKEN")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("empty SCW_SECRET_KEY")
	}

	parsedRegion, err := utils.ParseRegion(region)
	if err != nil {
		return nil, err
	}

	conf := &Config{
		DefaultProjectID: projectId,
		SecretKey:        secretKey,
		DefaultRegion:    parsedRegion,
	}

	// configures a default client for the region, using the above env vars
	client, err := conf.GetDeprecatedClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %#v", err)
	}

	return client, nil
}
