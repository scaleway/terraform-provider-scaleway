package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalewayToken_importBasic(t *testing.T) {
	resourceName := "scaleway_token.base"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayTokenConfig_Update,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
