package baremetal_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	baremetalSDK "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal"
)

const (
	OfferName = "EM-I215E-NVME"
	Zone      = "fr-par-2"
	OfferID   = "ddb09834-8680-4339-b5e3-19ca5205fa1f"
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
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "iridium"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "AMD EPYC 7313P"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "16"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "3000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "32"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1920383410176"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1920383410176"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "256000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "3200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc", "true"),
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
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "iridium"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "AMD EPYC 7313P"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "16"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "3000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "32"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1920383410176"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1920383410176"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "256000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "3200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc", "true"),
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
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "iridium"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "AMD EPYC 7313P"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "16"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "3000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "32"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1920383410176"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1920383410176"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "256000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "3200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc", "true"),
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
