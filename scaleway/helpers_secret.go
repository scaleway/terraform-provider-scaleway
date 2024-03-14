package scaleway

import (
	"encoding/base64"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultSecretTimeout = 5 * time.Minute
)

// secretAPIWithRegion returns a new Secret API and the region for a Create request
func secretAPIWithRegion(d *schema.ResourceData, m interface{}) (*secret.API, scw.Region, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// secretAPIWithRegionAndDefault returns a new Secret API and the region for a Create request
func secretAPIWithRegionAndDefault(d *schema.ResourceData, m interface{}, defaultRegion scw.Region) (*secret.API, scw.Region, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegionWithDefault(d, m, defaultRegion)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// secretAPIWithRegionAndProjectID returns a new Secret API, with region and projectID
func secretAPIWithRegionAndProjectID(d *schema.ResourceData, m interface{}) (*secret.API, scw.Region, string, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", "", err
	}

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, projectID, nil
}

// SecretAPIWithRegionAndID returns a Secret API with locality and ID extracted from the state
func SecretAPIWithRegionAndID(m interface{}, id string) (*secret.API, scw.Region, string, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

// SecretVersionAPIWithRegionAndID returns a Secret API with locality and Nested ID extracted from the state
func SecretVersionAPIWithRegionAndID(m interface{}, id string) (*secret.API, scw.Region, string, string, error) {
	region, id, revision, err := locality.ParseLocalizedNestedID(id)
	if err != nil {
		return nil, "", "", "", err
	}
	api := secret.NewAPI(meta.ExtractScwClient(m))
	return api, scw.Region(region), id, revision, nil
}

func isBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}

func Base64Encoded(data []byte) string {
	if isBase64Encoded(data) {
		return string(data)
	}
	return base64.StdEncoding.EncodeToString(data)
}
