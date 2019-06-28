package scaleway

import (
	"fmt"
	"testing"

	"gitlab.infra.online.net/front/protobuf/scaleway-sdk-go/api/instance/v1"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

// Check that reverse is handled at creation and update time
var testAccScalewayComputeInstanceSecurityGroupConfig = []string{
	`
		resource "scaleway_compute_instance_security_group" "base" {
			name = "sg-name"

			inbound_rule {
				port_range = "22"
				ip_range = "0.0.0.0"
            }
			inbound_rule {
				port_range = "0-1024"
				ip_range = "8.8.8.8"
            }
		}
	`,
	`
		resource "scaleway_compute_instance_security_group" "base" {
			inbound_default_policy = "accept"
			outbound_default_policy = "drop"
		}
	`,
}

func TestAccScalewayComputeInstanceSecurityGroup(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayComputeInstanceSecurityGroupConfig[0],
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						fmt.Println(s)
						return nil
					},
					testAccCheckScalewayComputeInstanceSecurityGroupExists("scaleway_compute_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_default_policy", "drop"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "outbound_default_policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_rule.#", "1"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_rule.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_rule.0.port_range", "22-22"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_rule.0.ip_range", "0.0.0.0/32"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_rule.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_rule.1.port_range", "0-1024"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_rule.1.ip_range", "8.8.8.8/32"),
				),
			},
			{
				Config: testAccScalewayComputeInstanceSecurityGroupConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceSecurityGroupExists("scaleway_compute_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_default_policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "outbound_default_policy", "drop"),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceSecurityGroupExists(n string) resource.TestCheckFunc {
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
		_, err = instanceApi.GetSecurityGroup(&instance.GetSecurityGroupRequest{
			SecurityGroupID: ID,
			Zone:            zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayComputeInstanceSecurityGroupDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta().(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_compute_instance_security_group" {
			continue
		}

		zone, ID, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetSecurityGroup(&instance.GetSecurityGroupRequest{
			Zone:            zone,
			SecurityGroupID: ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("security group (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
