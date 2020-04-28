package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestAccScalewayDataSourceBaremetalOffer_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalOfferExists("data.scaleway_baremetal_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test1", "name", "HC-BM1-L"),
					testAccCheckScalewayBaremetalOfferExists("data.scaleway_baremetal_offer.test2"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "offer_id", "fr-par-2/3ab0dc29-2fd4-486e-88bf-d08fbf49214b"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "name", "HC-BM1-L"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "include_disabled", "false"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "bandwidth", "1000"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "commercial_range", "high_cpu"),
					//resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "stock", "available"), // skipping this as stocks vary too much
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.name", "Intel Xeon Gold 5120"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.core_count", "28"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.frequency", "2200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.0.thread_count", "56"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.name", "Intel Xeon Gold 5120"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.core_count", "28"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.frequency", "2200"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "cpu.1.thread_count", "56"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.0.capacity", "1024"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.1.capacity", "1024"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.2.type", "NVMe"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "disk.3.capacity", "1024"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.type", "DDR4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.capacity", "384"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.frequency", "2133"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test2", "memory.0.ecc", "true"),
					testAccCheckScalewayBaremetalOfferExists("data.scaleway_baremetal_offer.test3"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_offer.test3", "name", "HC-BM1-L"),
				),
			},
		},
	})
}

func testAccCheckScalewayBaremetalOfferExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		fmt.Println(rs)

		zone, id, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		meta := testAccProvider.Meta().(*Meta)
		baremetalApi := baremetal.NewAPI(meta.scwClient)
		resp, err := baremetalApi.ListOffers(&baremetal.ListOffersRequest{
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
