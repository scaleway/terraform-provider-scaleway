package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	api "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sweepZones(region string, f func(scwClient *scw.Client) error) error {
	scwRegion, err := scw.ParseRegion(region)
	if err != nil {
		return err
	}
	for _, zone := range scwRegion.GetZones() {
		client, err := sharedClientForZone(zone.String())
		if err != nil {
			return err
		}
		err = f(client)
		if err != nil {
			l.Warningf("error running sweepZones, ignoring: %s", err)
		}
	}
	return nil
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

// sharedClientForZone returns a Scaleway client needed for the sweeper
// functions for a given zone {fr-par-1,fr-par-2,nl-ams-1}
func sharedClientForZone(zone string) (*scw.Client, error) {
	scwZone, err := scw.ParseZone(zone)
	if err != nil {
		return nil, err
	}
	scwRegion, err := scwZone.Region()
	if err != nil {
		return nil, err
	}

	meta, err := buildTestConfigForTests(scwRegion.String())
	if err != nil {
		return nil, err
	}

	meta.DefaultZone = scwZone

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

	accessKey := os.Getenv("SCW_ACCESS_KEY")
	if accessKey == "" {
		return nil, fmt.Errorf("empty SCW_ACCESS_KEY")
	}
	parsedRegion, err := scw.ParseRegion(region)
	if err != nil {
		return nil, err
	}

	return &Meta{
		DefaultOrganizationID: organizationID,
		SecretKey:             secretKey,
		AccessKey:             accessKey,
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
