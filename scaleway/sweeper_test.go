package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/minio/minio-go/v6"
	api "github.com/nicolai86/scaleway-sdk"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedClientForRegion returns a common scaleway client needed for the sweeper
// functions for a given region {par1,ams1}
func sharedDeprecatedClientForRegion(region string) (*api.API, error) {
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
	client, err := conf.GetDeprecatedClient()
	if err != nil {
		return nil, fmt.Errorf("error getting Scaleway client: %#v", err)
	}

	return client, nil
}

// sharedS3ClientForRegion returns a common S3 client needed for the sweeper
func sharedS3ClientForRegion(region string) (*minio.Client, error) {

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
	client, err := conf.GetS3Client()
	if err != nil {
		return nil, fmt.Errorf("error getting S3 client: %#v", err)
	}

	return client, nil

}
