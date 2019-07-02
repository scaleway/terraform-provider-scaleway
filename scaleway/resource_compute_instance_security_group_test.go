package scaleway

import (
	"fmt"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

// Check that reverse is handled at creation and update time
var testAccScalewayComputeInstanceSecurityGroupConfig = []string{
	`
		resource "scaleway_compute_instance_security_group" "base" {
			name = "sg-name"

			rule {
				port_range = "22"
				ip_range = "0.0.0.0"
            }

			rule {
				port_range = "1-1024"
				ip_range = "8.8.8.0/24"
			}
			
			rule {
				type = "outbound"
				port_range = "3000"
				ip_range = "0.0.0.0"
			}
		}
	`,
	`
		resource "scaleway_compute_instance_security_group" "base" {
			inbound_default_policy = "accept"
			outbound_default_policy = "drop"

			rule {
				port_range = "22"
				ip_range = "0.0.0.0"
			}

			rule {
				port_range = "1-1024"
				ip_range = "8.8.8.0/24"
			}
			
			rule {
				type = "outbound"
				port_range = "3000"
				ip_range = "0.0.0.0"
			}
			
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
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.#", "3"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.4082861517.type", "inbound"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.4082861517.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.4082861517.port_range", "22"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.4082861517.ip_range", "0.0.0.0"),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleIs("scaleway_compute_instance_security_group.base", 4082861517, "accept"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.type", "inbound"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.port_range", "1-1024"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.ip_range", "8.8.8.0/24"),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleIs("scaleway_compute_instance_security_group.base", 1078896110, "accept"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.type", "outbound"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.port_range", "3000"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.ip_range", "0.0.0.0"),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleIs("scaleway_compute_instance_security_group.base", 813120688, "drop"),
				),
			},
			{
				Config: testAccScalewayComputeInstanceSecurityGroupConfig[1],
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						fmt.Println(s)
						return nil
					},
					testAccCheckScalewayComputeInstanceSecurityGroupExists("scaleway_compute_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "inbound_default_policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "outbound_default_policy", "drop"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.#", "3"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.4082861517.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.4082861517.port_range", "22"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.4082861517.ip_range", "0.0.0.0"),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleIs("scaleway_compute_instance_security_group.base", 4082861517, "drop"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.type", "inbound"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.port_range", "1-1024"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.1078896110.ip_range", "8.8.8.0/24"),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleIs("scaleway_compute_instance_security_group.base", 1078896110, "drop"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.type", "outbound"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.port_range", "3000"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_security_group.base", "rule.813120688.ip_range", "0.0.0.0"),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleIs("scaleway_compute_instance_security_group.base", 813120688, "accept"),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceSecurityGroupRuleIs(name string, key int, action instance.SecurityGroupRuleAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		res, err := instanceApi.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
			SecurityGroupID: ID,
			Zone:            zone,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		for _, rule := range res.Rules {
			flat := securityGroupRuleFlatten(rule)
			if securityGroupRuleHash(flat) == key {
				if rule.Action != action {
					return fmt.Errorf("rule with hash %d shoud have action %s got %s", key, action, rule.Action)
				}
				return nil
			}
		}

		return fmt.Errorf("could not find a rule with hash %d", key)
	}
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
