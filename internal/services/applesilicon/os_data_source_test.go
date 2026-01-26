package applesilicon_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

var (
	ErrAppleSiliconOSNotFound = errors.New("not found")
)

func TestAccDataSourceOS_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_apple_silicon_os" "by_name" {
						name = "macos-ventura-13.6"
					}
					
					data "scaleway_apple_silicon_os" "by_id" {
						os_id = "cafecafe-5018-4dcd-bd08-35f031b0ac3e"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppleSiliconOsExists(tt, "data.scaleway_apple_silicon_os.by_id"),
					testAccCheckAppleSiliconOsExists(tt, "data.scaleway_apple_silicon_os.by_name"),

					resource.TestCheckResourceAttrSet("data.scaleway_apple_silicon_os.by_name", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_apple_silicon_os.by_name", "version"),
					resource.TestCheckResourceAttrSet("data.scaleway_apple_silicon_os.by_name", "os_id"),

					resource.TestCheckResourceAttrSet("data.scaleway_apple_silicon_os.by_id", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_apple_silicon_os.by_id", "version"),
					resource.TestCheckResourceAttrSet("data.scaleway_apple_silicon_os.by_id", "os_id"),
				),
			},
		},
	})
}

func testAccCheckAppleSiliconOsExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("%w: %s", ErrAppleSiliconOSNotFound, n)
		}

		zone, ID, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		applesiliconAPI := applesilicon.NewAPI(tt.Meta.ScwClient())

		_, err = applesiliconAPI.GetOS(&applesilicon.GetOSRequest{
			OsID: ID,
			Zone: zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
