package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
						name = "HC-BM1-L"
					}
					
					data "scaleway_baremetal_offer" "test2" {
						zone     = "fr-par-2"
						offer_id = "3ab0dc29-2fd4-486e-88bf-d08fbf49214b"
					}
					
					data "scaleway_baremetal_offer" "test3" {
						offer_id = "fr-par-2/3ab0dc29-2fd4-486e-88bf-d08fbf49214b"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", "HC-BM1-L"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "offer_id", "fr-par-2/3ab0dc29-2fd4-486e-88bf-d08fbf49214b"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", "HC-BM1-L"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "high_cpu"),
					//resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "stock", "available"), // skipping this as stocks vary too much
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "Intel Xeon Gold 5120"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "14"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "2200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "28"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.name", "Intel Xeon Gold 5120"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.core_count", "14"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.frequency", "2200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.thread_count", "28"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1024000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1024000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.2.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.3.capacity", "1024000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "384000000000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "2133"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.is_ecc", "true"),
					testAccCheckScalewayBaremetalOfferExists(tt, "data.scaleway_baremetal_offer.test3"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test3", "name", "HC-BM1-L"),
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
		resp, err := baremetalAPI.ListOffers(&baremetal.ListOffersRequest{
			Zone: zone,
		}, scw.WithAllPages())

		if err != nil {
			return err
		}
		for _, offer := range resp.Offers {
			if offer.ID == id {
				return nil
			}
		}
		return fmt.Errorf("offer %s not found in zone %s", id, zone)
	}
}
