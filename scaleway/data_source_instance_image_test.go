package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayDataSourceInstanceImage_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_instance_image" "test" {
						image_id = "fr-par-1/cf44b8f5-77e2-42ed-8f1e-09ed5bb028fc"
					}

					data "scaleway_instance_image" "test2" {
						image_id = "cf44b8f5-77e2-42ed-8f1e-09ed5bb028fc"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "data.scaleway_instance_image.test"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test", "name", "Ubuntu 20.04 Focal Fossa"),

					testAccCheckScalewayInstanceImageExists(tt, "data.scaleway_instance_image.test2"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "image_id", "fr-par-1/cf44b8f5-77e2-42ed-8f1e-09ed5bb028fc"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "name", "Ubuntu 20.04 Focal Fossa"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "creation_date", "2020-07-09T10:38:16Z"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "modification_date", "2020-07-09T10:38:16Z"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "latest", "true"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "public", "true"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "from_server_id", ""),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "state", "available"),
					resource.TestCheckResourceAttr("data.scaleway_instance_image.test2", "root_volume_id", "6e66445c-e52e-4cfa-bf4c-f36e291e2c30"),
					resource.TestCheckNoResourceAttr("data.scaleway_instance_image.test2", "additional_volume_ids"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceImageExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, ID, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		instanceAPI := instance.NewAPI(tt.Meta.scwClient)
		_, err = instanceAPI.GetImage(&instance.GetImageRequest{
			ImageID: ID,
			Zone:    zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}
