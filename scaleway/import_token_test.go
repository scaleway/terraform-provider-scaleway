package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalewayToken_importBasic(t *testing.T) {
	t.Parallel()

	resourceName := "scaleway_token.base"

	resource.Test(t, resource.TestCase{
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
