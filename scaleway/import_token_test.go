package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalewayToken_importBasic(t *testing.T) {
	resourceName := "scaleway_token.base"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayTokenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayTokenConfig_Update,
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
