package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewaySecurityGroup_importBasic(t *testing.T) {
	resourceName := "scaleway_security_group.base"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewaySecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewaySecurityGroupConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
