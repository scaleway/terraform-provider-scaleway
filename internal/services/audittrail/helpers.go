package audittrail

import (
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

	return api, region, nil
}
