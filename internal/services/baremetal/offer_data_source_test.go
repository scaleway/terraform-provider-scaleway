package baremetal_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	baremetalSDK "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal"
)

var (
	OfferName = getenv("OFFER_NAME", "EM-I215E-NVME")
	Zone      = getenv("ZONE", "fr-par-2")
)

func TestAccDataSourceOffer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_offer" "test1" {
						zone = "%s"
						name = "%s"
					}
					
					data "scaleway_baremetal_offer" "test2" {
						offer_id = data.scaleway_baremetal_offer.test1.offer_id
					}
				`, Zone, OfferName),
				Check: resource.ComposeTestCheckFunc(
					isOfferPresent(tt, "data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", OfferName),
					isOfferPresent(tt, "data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttrPair("data.scaleway_baremetal_offer.test2", "offer_id", "data.scaleway_baremetal_offer.test1", "offer_id"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", OfferName),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "commercial_range"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "include_disabled"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "bandwidth"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.core_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.frequency"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.0.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.0.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.1.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.1.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.frequency"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc"),
				),
			},
		},
	})
}

func TestAccDataSourceOffer_SubscriptionPeriodHourly(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_offer" "test1" {
						zone = "%s"
						name = "%s"

						subscription_period = "hourly"
					}
					
					data "scaleway_baremetal_offer" "test2" {
						offer_id = data.scaleway_baremetal_offer.test1.offer_id
					}
				`, Zone, OfferName),
				Check: resource.ComposeTestCheckFunc(
					isOfferPresent(tt, "data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", OfferName),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "subscription_period", "hourly"),
					isOfferPresent(tt, "data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttrPair("data.scaleway_baremetal_offer.test2", "offer_id", "data.scaleway_baremetal_offer.test1", "offer_id"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", OfferName),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "subscription_period", "hourly"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "commercial_range"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "include_disabled"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "bandwidth"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.core_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.frequency"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.0.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.0.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.1.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.1.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.frequency"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc"),
				),
			},
		},
	})
}

func TestAccDataSourceOffer_SubscriptionPeriodMonthly(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_offer" "test1" {
						zone = "%s"
						name = "%s"

						subscription_period = "monthly"
					}
					
					data "scaleway_baremetal_offer" "test2" {
						offer_id = data.scaleway_baremetal_offer.test1.offer_id
					}
				`, Zone, OfferName),
				Check: resource.ComposeTestCheckFunc(
					isOfferPresent(tt, "data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", OfferName),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "subscription_period", "monthly"),
					isOfferPresent(tt, "data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttrPair("data.scaleway_baremetal_offer.test2", "offer_id", "data.scaleway_baremetal_offer.test1", "offer_id"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", OfferName),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "subscription_period", "monthly"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "commercial_range"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "include_disabled"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "bandwidth"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.core_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.frequency"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.0.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.0.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.1.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "disk.1.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.type"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.capacity"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.frequency"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc"),
				),
			},
		},
	})
}

func isOfferPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, id, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		api := baremetalSDK.NewAPI(tt.Meta.ScwClient())

		_, err = baremetal.FindOfferByID(context.Background(), api, zone, id)
		if err != nil {
			return err
		}

		return nil
	}
}
