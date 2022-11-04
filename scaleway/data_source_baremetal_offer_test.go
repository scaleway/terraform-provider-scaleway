package scaleway

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
)

func TestAccScalewayDataSourceBaremetalOffer_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_baremetal_offer" "test1" {
						zone = "fr-par-2"
						name = "EM-A210R-HDD"
					}
					
					data "scaleway_baremetal_offer" "test2" {
						zone     = "fr-par-2"
						offer_id = "25dcf38b-c90c-4b18-97a2-6956e9d1e113"
					}
					
					data "scaleway_baremetal_offer" "test3" {
						offer_id = "fr-par-2/25dcf38b-c90c-4b18-97a2-6956e9d1e113"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", "EM-A210R-HDD"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "offer_id", "fr-par-2/25dcf38b-c90c-4b18-97a2-6956e9d1e113"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", "EM-A210R-HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "aluminium"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "aluminium"),
					// resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "stock", "available"), // skipping this as stocks vary too much
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "AMD Ryzen PRO 3600"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "6"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "3600"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "12"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1000000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1000000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "16000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "3200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc", "true"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test3"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test3", "name", "EM-A210R-HDD"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceBaremetalOffer_SubscriptionPeriodHourly(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_baremetal_offer" "test1" {
						zone = "fr-par-2"
						name = "EM-A210R-HDD"

						subscription_period = "hourly"
					}
					
					data "scaleway_baremetal_offer" "test2" {
						zone     = "fr-par-2"
						offer_id = "25dcf38b-c90c-4b18-97a2-6956e9d1e113"
					}
					
					data "scaleway_baremetal_offer" "test3" {
						offer_id = "fr-par-2/25dcf38b-c90c-4b18-97a2-6956e9d1e113"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", "EM-A210R-HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "subscription_period", "hourly"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "offer_id", "fr-par-2/25dcf38b-c90c-4b18-97a2-6956e9d1e113"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", "EM-A210R-HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "subscription_period", "hourly"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "aluminium"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "aluminium"),
					// resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "stock", "available"), // skipping this as stocks vary too much
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "AMD Ryzen PRO 3600"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "6"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "3600"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "12"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1000000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1000000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "16000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "3200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc", "true"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test3"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test3", "name", "EM-A210R-HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test3", "subscription_period", "hourly"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceBaremetalOffer_SubscriptionPeriodMonthly(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_baremetal_offer" "test1" {
						zone = "fr-par-2"
						name = "EM-A210R-HDD"

						subscription_period = "monthly"
					}
					
					data "scaleway_baremetal_offer" "test2" {
						zone     = "fr-par-2"
						offer_id = "10b2832f-eecd-46c5-9812-014125a215c8"
					}
					
					data "scaleway_baremetal_offer" "test3" {
						offer_id = "fr-par-2/10b2832f-eecd-46c5-9812-014125a215c8"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", "EM-A210R-HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "subscription_period", "monthly"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "offer_id", "fr-par-2/10b2832f-eecd-46c5-9812-014125a215c8"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", "EM-A210R-HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "subscription_period", "monthly"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "aluminium"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "aluminium"),
					// resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "stock", "available"), // skipping this as stocks vary too much
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "AMD Ryzen PRO 3600"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "6"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "3600"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "12"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1000000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1000000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "16000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "3200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc", "true"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test3"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test3", "name", "EM-A210R-HDD"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test3", "subscription_period", "monthly"),
				),
			},
		},
	})
}

func testAccCheckScalewayBaremetalOfferExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, id, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		baremetalAPI := baremetal.NewAPI(tt.Meta.scwClient)
		_, err = baremetalFindOfferById(baremetalAPI, zone, id, context.Background())
		if err != nil {
			return err
		}

		return nil
	}
}
