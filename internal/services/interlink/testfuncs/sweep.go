package interlinktestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_interlink_routing_policy", &resource.Sweeper{
		Name: "scaleway_interlink_routing_policy",
		F:    testSweepRoutingPolicy,
	})
}

func testSweepRoutingPolicy(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		interlinkAPI := interlink.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the interlink routing policies in (%s)", region)

		listPolicies, err := interlinkAPI.ListRoutingPolicies(&interlink.ListRoutingPoliciesRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing interlink routing policies in (%s) in sweeper: %w", region, err)
		}

		for _, policy := range listPolicies.RoutingPolicies {
			err := interlinkAPI.DeleteRoutingPolicy(&interlink.DeleteRoutingPolicyRequest{
				Region:          region,
				RoutingPolicyID: policy.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting interlink routing policy in sweeper: %w", err)
			}
		}

		return nil
	})
}
