package audittrail

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func CheckEventsOccurrence(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return errors.New("not found: " + resourceName)
		}

		countStr := rs.Primary.Attributes["events.#"]

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return fmt.Errorf("could not parse events.# as integer: %w", err)
		}

		if count != 1 {
			return fmt.Errorf("expected exactly 1 event, got %d", count)
		}

		return nil
	}
}
