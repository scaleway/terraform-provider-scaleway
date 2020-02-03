package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_security_group_rules", &resource.Sweeper{
		Name: "scaleway_instance_security_group_rules",
		F:    testSweepComputeInstanceSecurityGroupRules,
	})
}

func TestAccScalewayInstanceSecurityGroupRules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: conf1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("scaleway_instance_security_group.sg01"),
				),
			},
		},
	})
}

var conf1 = `
resource scaleway_instance_server s01 { 
	security_group_id = scaleway_instance_security_group.sg01.id
}

resource scaleway_instance_security_group sg01 {
	external_rules = true
}

resource scaleway_instance_security_group_rules sgrs01 {
	security_group_id = scaleway_instance_security_group.sg01.id
	inbound_rule {
		address = scaleway_instance_server.s01.private_ip
	}
}
`

func testSweepComputeInstanceSecurityGroupRules(region string) error {
	return nil
}
