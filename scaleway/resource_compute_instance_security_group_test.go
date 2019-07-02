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
				port = "22"
				ip = "10.10.10.10"
            }

			rule {
				port_range = "1-1024"
				ip_range = "8.8.8.0/24"
			}
			
			rule {
				type = "outbound"
				port_range = "3000"
			}
		}
	`,
	`
		resource "scaleway_compute_instance_security_group" "base" {
			inbound_default_policy = "accept"
			outbound_default_policy = "drop"

			rule {
				port = "22"
				ip = "10.10.10.10"
			}

			rule {
				port_range = "1-1024"
				ip_range = "8.8.8.0/24"
			}
			
			rule {
				type = "outbound"
				port_range = "3000"
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
					testAccCheckScalewayComputeInstanceSecurityGroupRuleMatch("scaleway_compute_instance_security_group.base", 2421519970,
						instance.SecurityGroupRuleDirectionInbound, instance.SecurityGroupRuleActionAccept, instance.SecurityGroupRuleProtocolTCP, "10.10.10.10", 22, 0),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleMatch("scaleway_compute_instance_security_group.base", 2236276624,
						instance.SecurityGroupRuleDirectionInbound, instance.SecurityGroupRuleActionAccept, instance.SecurityGroupRuleProtocolTCP, "8.8.8.0/24", 1, 1024),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleMatch("scaleway_compute_instance_security_group.base", 3624953052,
						instance.SecurityGroupRuleDirectionOutbound, instance.SecurityGroupRuleActionDrop, instance.SecurityGroupRuleProtocolTCP, "0.0.0.0/0", 3000, 0),
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
					testAccCheckScalewayComputeInstanceSecurityGroupRuleMatch("scaleway_compute_instance_security_group.base", 2421519970,
						instance.SecurityGroupRuleDirectionInbound, instance.SecurityGroupRuleActionDrop, instance.SecurityGroupRuleProtocolTCP, "10.10.10.10", 22, 0),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleMatch("scaleway_compute_instance_security_group.base", 2236276624,
						instance.SecurityGroupRuleDirectionInbound, instance.SecurityGroupRuleActionDrop, instance.SecurityGroupRuleProtocolTCP, "8.8.8.0/24", 1, 1024),
					testAccCheckScalewayComputeInstanceSecurityGroupRuleMatch("scaleway_compute_instance_security_group.base", 3624953052,
						instance.SecurityGroupRuleDirectionOutbound, instance.SecurityGroupRuleActionAccept, instance.SecurityGroupRuleProtocolTCP, "0.0.0.0/0", 3000, 0),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceSecurityGroupRuleMatch(name string, key int, direction instance.SecurityGroupRuleDirection, action instance.SecurityGroupRuleAction, protocol instance.SecurityGroupRuleProtocol, ipRange string, portFrom uint32, portTo uint32) resource.TestCheckFunc {
	return testAccCheckScalewayComputeInstanceSecurityGroupRuleIs(name, key, func(rule *instance.SecurityGroupRule) error {
		if rule.Direction != direction {
			return fmt.Errorf("direction with hash %d shoud be %s got %s", key, direction, rule.Direction)
		}
		if rule.Action != action {
			return fmt.Errorf("rule with hash %d shoud be %s got %s", key, action, rule.Action)
		}
		if rule.Protocol != protocol {
			return fmt.Errorf("protocol with hash %d shoud be %s got %s", key, protocol, rule.Protocol)
		}
		if rule.IPRange != ipRange {
			return fmt.Errorf("ip_range with hash %d shoud be %s got %s", key, ipRange, rule.IPRange)
		}
		if rule.DestPortFrom != portFrom {
			return fmt.Errorf("dest_port_from with hash %d shoud be %d got %d", key, portFrom, rule.DestPortFrom)
		}
		if rule.DestPortTo != portTo {
			return fmt.Errorf("dest_port_to with hash %d shoud be %d got %d", key, portTo, rule.DestPortTo)
		}
		return nil
	})
}

func testAccCheckScalewayComputeInstanceSecurityGroupRuleIs(name string, key int, test func(rule *instance.SecurityGroupRule) error) resource.TestCheckFunc {
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
				return test(rule)
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
