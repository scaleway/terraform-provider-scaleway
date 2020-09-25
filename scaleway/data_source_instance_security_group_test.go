package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceInstanceSecurityGroup_Basic(t *testing.T) {
	securityGroupName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "scaleway_instance_security_group" "main" {
	name 	   = "` + securityGroupName + `"

}

data "scaleway_instance_security_group" "prod" {
	name = "${scaleway_instance_security_group.main.name}"
}

data "scaleway_instance_security_group" "stg" {
	security_group_id = "${scaleway_instance_security_group.main.id}"
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("data.scaleway_instance_security_group.prod"),
					resource.TestCheckResourceAttr("data.scaleway_instance_security_group.prod", "name", securityGroupName),
					testAccCheckScalewayInstanceSecurityGroupExists("data.scaleway_instance_security_group.stg"),
					resource.TestCheckResourceAttr("data.scaleway_instance_security_group.stg", "name", securityGroupName),
				),
			},
			{
				Config: `
resource "scaleway_instance_security_group" "main" {
	name 	   = "` + securityGroupName + `"
}

data "scaleway_instance_security_group" "prod" {
	security_group_id = "${scaleway_instance_security_group.main.id}"
}

data "scaleway_instance_security_group" "stg" {
	name = "${scaleway_instance_security_group.main.name}"
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("data.scaleway_instance_security_group.prod"),
					resource.TestCheckResourceAttr("data.scaleway_instance_security_group.prod", "name", securityGroupName),
					testAccCheckScalewayInstanceSecurityGroupExists("data.scaleway_instance_security_group.stg"),
					resource.TestCheckResourceAttr("data.scaleway_instance_security_group.stg", "name", securityGroupName),
				),
			},
		},
	})
}
