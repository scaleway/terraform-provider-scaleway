package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayDataSourceInstanceImage_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "scaleway_instance_image" "prod" {
	name = "golang 1.10"
}

data "scaleway_instance_image" "stg" {
	image_id = "43213956-c7a3-44b8-9d96-d51fa7457969"
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists("data.scaleway_instance_image.prod"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.prod", "name", "Golang 1.10"),
					testAccCheckScalewayInstanceImageExists("data.scaleway_instance_image.stg"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "name", "Golang 1.10"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "creation_date", "2018-04-12T10:22:46Z"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "modification_date", "2018-04-12T15:02:26Z"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "latest", "true"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "public", "true"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "from_server_id", ""),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "state", "available"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "default_bootscript_id", "b1e68c26-a19c-4eac-9222-498b22bd7ad9"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.stg", "root_volume_id", "8fa97c03-ca3b-4267-ba19-2d38190b1c82"),
					resource.TestCheckNoResourceAttr("data.scaleway_instance_image.stg", "additional_volume_ids"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceImageExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, ID, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		meta := testAccProvider.Meta().(*Meta)
		instanceApi := instance.NewAPI(meta.scwClient)
		_, err = instanceApi.GetImage(&instance.GetImageRequest{
			ImageID: ID,
			Zone:    zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}
