package audittrail

import (
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	audittrailSDK "github.com/scaleway/scaleway-sdk-go/api/audit_trail/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// newAPIWithRegionAndProjectID returns a new Audit Trail API, with region and projectID
func newAPIWithRegion(d *schema.ResourceData, m any) (*audittrailSDK.API, scw.Region, error) {
	api := audittrailSDK.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	// In audit trail sdk-go so far only fr-par and nl-ams are supported
	if !slices.Contains(api.Regions(), region) {
		return nil, "", fmt.Errorf("invalid api region, expected one of %s, got: %s", api.Regions(), region)
	}

	return api, region, nil
}
