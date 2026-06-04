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
	resource.AddTestSweepers("scaleway_interlink_link", &resource.Sweeper{
		Name: "scaleway_interlink_link",
		F:    testSweepLink,
	})

	resource.AddTestSweepers("scaleway_interlink_routing_policy", &resource.Sweeper{
		Name:         "scaleway_interlink_routing_policy",
		F:            testSweepRoutingPolicy,
		Dependencies: []string{"scaleway_interlink_link"},
	})
}

func testSweepLink(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		interlinkAPI := interlink.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the interlink links in (%s)", region)

		listLinks, err := interlinkAPI.ListLinks(&interlink.ListLinksRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing interlink links in (%s) in sweeper: %w", region, err)
		}

		for _, link := range listLinks.Links {
			_, err := interlinkAPI.DeleteLink(&interlink.DeleteLinkRequest{
				Region: region,
				LinkID: link.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting interlink link in sweeper: %w", err)
			}
		}

		return nil
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
