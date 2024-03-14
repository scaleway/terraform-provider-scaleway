package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func init() {
	resource.AddTestSweepers("scaleway_webhosting", &resource.Sweeper{
		Name: "scaleway_webhosting",
		F:    testSweepWebhosting,
	})
}

func testSweepWebhosting(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		webhsotingAPI := webhosting.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting the hostings in (%s)", region)

		listHostings, err := webhsotingAPI.ListHostings(&webhosting.ListHostingsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing hostings in (%s) in sweeper: %s", region, err)
		}

		for _, hosting := range listHostings.Hostings {
			_, err := webhsotingAPI.DeleteHosting(&webhosting.DeleteHostingRequest{
				HostingID: hosting.ID,
				Region:    region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting hosting in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayWebhosting_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayWebhostingDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				data "scaleway_webhosting_offer" "by_name" {
				  name = "lite"
				}

				resource "scaleway_webhosting" "main" {
				  offer_id     = data.scaleway_webhosting_offer.by_name.offer_id
				  email        = "hashicorp@scaleway.com"
				  domain       = "scaleway.com"
				  tags         = ["devtools", "provider", "terraform"]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayWebhostingExists(tt, "scaleway_webhosting.main"),
					resource.TestCheckResourceAttrPair("scaleway_webhosting.main", "offer_id", "data.scaleway_webhosting_offer.by_name", "offer_id"),
					resource.TestCheckResourceAttr("scaleway_webhosting.main", "email", "hashicorp@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_webhosting.main", "domain", "scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_webhosting.main", "status", webhosting.HostingStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_webhosting.main", "tags.0", "devtools"),
					resource.TestCheckResourceAttr("scaleway_webhosting.main", "tags.1", "provider"),
					resource.TestCheckResourceAttr("scaleway_webhosting.main", "tags.2", "terraform"),
					resource.TestCheckResourceAttrSet("scaleway_webhosting.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_webhosting.main", "created_at"),
					testCheckResourceAttrUUID("scaleway_webhosting.main", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayWebhostingExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := scaleway.WebhostingAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetHosting(&webhosting.GetHostingRequest{
			HostingID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayWebhostingDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_webhosting" {
				continue
			}

			api, region, id, err := scaleway.WebhostingAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			res, err := api.GetHosting(&webhosting.GetHostingRequest{
				HostingID: id,
				Region:    region,
			})

			if err == nil && res.Status != webhosting.HostingStatusUnknownStatus {
				return fmt.Errorf("hosting (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
