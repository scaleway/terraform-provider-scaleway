package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
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
	config, err := buildTestConfigForTests(region)
	if err != nil {
		return nil, err
	}

	// configures a default client for the region, using the above env vars
	client, err := config.GetDeprecatedClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %#v", err)
	}

	return client, nil
}

// sharedClientForRegion returns a Scaleway client needed for the sweeper
// functions for a given region {fr-par,nl-ams}
func sharedClientForRegion(region string) (*scw.Client, error) {

	config, err := buildTestConfigForTests(region)
	if err != nil {
		return nil, err
	}

	// configures a default client for the region, using the above env vars
	client, err := config.GetScwClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %s", err)
	}

	return client, nil
}

// buildTestConfigForTests creates a Config objects based on the region
// and the config variables needed for testing.
func buildTestConfigForTests(region string) (*Config, error) {
	projectID := os.Getenv("SCW_DEFAULT_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("SCALEWAY_ORGANIZATION")
	}
	if projectID == "" {
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

	return &Config{
		DefaultProjectID: projectID,
		SecretKey:        secretKey,
		DefaultRegion:    parsedRegion,
	}, nil
}

// sharedS3ClientForRegion returns a common S3 client needed for the sweeper
func sharedS3ClientForRegion(region string) (*s3.S3, error) {

	config, err := buildTestConfigForTests(region)
	if err != nil {
		return nil, err
	}

	// configures a default client for the region, using the above env vars
	client, err := config.GetS3Client()
	if err != nil {
		return nil, fmt.Errorf("error getting S3 client: %#v", err)
	}

	return client, nil

}
