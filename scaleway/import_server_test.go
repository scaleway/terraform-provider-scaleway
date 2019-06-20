package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalewayServer_importBasic(t *testing.T) {
	resourceName := "scaleway_server.base"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayServerConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"state_detail",
				},
			},
		},
	})
}
