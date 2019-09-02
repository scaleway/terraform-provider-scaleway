package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/resource"
	api "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedDeprecatedClientForRegion returns a scaleway deprecated client needed for the sweeper
// functions for a given region {par1,ams1}
func sharedDeprecatedClientForRegion(region string) (*api.API, error) {
	meta, err := buildTestConfigForTests(region)
	if err != nil {
		return nil, err
	}

	// configures a default client for the region, using the above env vars
	err = meta.bootstrapDeprecatedClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %#v", err)
	}

	return meta.deprecatedClient, nil
}

// sharedClientForRegion returns a Scaleway client needed for the sweeper
// functions for a given region {fr-par,nl-ams}
func sharedClientForRegion(region string) (*scw.Client, error) {
	meta, err := buildTestConfigForTests(region)
	if err != nil {
		return nil, err
	}

	// configures a default client for the region, using the above env vars
	err = meta.bootstrapScwClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %s", err)
	}

	return meta.scwClient, nil
}

// buildTestConfigForTests creates a Config objects based on the region
// and the config variables needed for testing.
func buildTestConfigForTests(region string) (*Meta, error) {
	organizationID := os.Getenv("SCW_DEFAULT_ORGANIZATION_ID")
	if organizationID == "" {
		organizationID = os.Getenv("SCALEWAY_ORGANIZATION")
	}
	if organizationID == "" {
		return nil, fmt.Errorf("empty SCW_DEFAULT_ORGANIZATION_ID")
	}

	secretKey := os.Getenv("SCW_SECRET_KEY")
	if secretKey == "" {
		secretKey = os.Getenv("SCALEWAY_TOKEN")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("empty SCW_SECRET_KEY")
	}

	parsedRegion, err := scw.ParseRegion(region)
	if err != nil {
		return nil, err
	}

	return &Meta{
		DefaultOrganizationID: organizationID,
		SecretKey:             secretKey,
		DefaultRegion:         parsedRegion,
	}, nil
}

// sharedS3ClientForRegion returns a common S3 client needed for the sweeper
func sharedS3ClientForRegion(region string) (*s3.S3, error) {

	meta, err := buildTestConfigForTests(region)
	if err != nil {
		return nil, err
	}

	// configures a default client for the region, using the above env vars
	err = meta.bootstrapS3Client()
	if err != nil {
		return nil, fmt.Errorf("error getting S3 client: %#v", err)
	}

	return meta.s3Client, nil

}
