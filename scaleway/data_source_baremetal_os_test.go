package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func TestAccScalewayDataSourceBaremetalOS_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_baremetal_os" "by_name" {
						name = "Ubuntu"
						version = "20.04 LTS (Focal Fossa)"
					}
					
					data "scaleway_baremetal_os" "by_id" {
						os_id = "03b7f4ba-a6a1-4305-984e-b54fafbf1681"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalOsExists(tt, "data.scaleway_baremetal_os.by_id"),
					testAccCheckScalewayBaremetalOsExists(tt, "data.scaleway_baremetal_os.by_name"),

					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_os.by_name", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_os.by_name", "version"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_os.by_name", "os_id"),

					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_os.by_id", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_os.by_id", "version"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_os.by_id", "os_id"),
				),
			},
		},
	})
}

func testAccCheckScalewayBaremetalOsExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, ID, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		baremetalAPI := baremetal.NewAPI(tt.Meta.ScwClient())
		_, err = baremetalAPI.GetOS(&baremetal.GetOSRequest{
			OsID: ID,
			Zone: zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
