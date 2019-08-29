package scaleway

import (
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_security_group", &resource.Sweeper{
		Name: "scaleway_instance_security_group",
		F:    testSweepComputeInstanceSecurityGroup,
	})
}

// Test that we can add / update / delete rules
var testAccScalewayInstanceSecurityGroupConfig = []string{
	`
		resource "scaleway_instance_security_group" "base" {
			name = "sg-name"
			inbound_default_policy = "drop"
			
			inbound_rule {
			   	action = "accept"
			   	port = 80
				ip_range = "0.0.0.0/0"
            }

			inbound_rule {
		   		action = "accept"
		   		port = 22
				ip = "1.1.1.1"
			}
		}
	`,
	`
		resource "scaleway_instance_security_group" "base" {
			name = "sg-name"
			inbound_default_policy = "accept"

			inbound_rule {
			   	action = "drop"
			   	port = 80
				ip = "8.8.8.8"
            }

			inbound_rule {
			   	action = "accept"
			   	port = 80
				ip_range = "0.0.0.0/0"
            }

			inbound_rule {
		   		action = "accept"
		   		port = 22
				ip = "1.1.1.1"
			}
			
		}
	`,
	`
		resource "scaleway_instance_security_group" "base" {
			name = "sg-name"
			inbound_default_policy = "accept"
		}
	`,
}

// Test that we can use ICMP protocol
var testAccScalewayInstanceSecurityGroupConfigICMP = []string{
	`
		resource "scaleway_instance_security_group" "base" {
			inbound_rule {
				action = "accept"
				port = 80
				ip_range = "0.0.0.0/0"
            }
		}
	`,
	`
		resource "scaleway_instance_security_group" "base" {
			inbound_rule {
				action = "drop"
				protocol = "ICMP"
				ip = "8.8.8.8"
            }
		}
	`,
}

// Test that we can omit port on a rule
var testAccScalewayInstanceSecurityGroupConfigNoPort = []string{
	`
		resource "scaleway_instance_security_group" "base" {
			inbound_rule {
				action = "accept"
				ip_range = "0.0.0.0/0"
            }
		}
	`,
}

func TestAccScalewayInstanceSecurityGroup(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayInstanceSecurityGroupConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("scaleway_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_default_policy", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "outbound_default_policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.ip_range", "0.0.0.0/0"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "0.0.0.0/0",
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.port", "22"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.ip", "1.1.1.1"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 1, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "1.1.1.1",
						DestPortFrom: scw.Uint32Ptr(22),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
				),
			},
			{
				Config: testAccScalewayInstanceSecurityGroupConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("scaleway_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_default_policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "outbound_default_policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.#", "3"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.ip", "8.8.8.8"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "8.8.8.8",
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionDrop,
					}),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.1.ip_range", "0.0.0.0/0"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 1, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "0.0.0.0/0",
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.2.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.2.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.2.port", "22"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.2.ip", "1.1.1.1"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 2, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "1.1.1.1",
						DestPortFrom: scw.Uint32Ptr(22),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
				),
			},
			{
				Config: testAccScalewayInstanceSecurityGroupConfig[2],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.#", "0"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSecurityGroupICMP(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayInstanceSecurityGroupConfigICMP[0],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.ip_range", "0.0.0.0/0"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "0.0.0.0/0",
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
				),
			},
			{
				Config: testAccScalewayInstanceSecurityGroupConfigICMP[1],
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.protocol", "ICMP"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.port", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_rule.0.ip", "8.8.8.8"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "8.8.8.8",
						DestPortFrom: nil,
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolICMP,
						Action:       instance.SecurityGroupRuleActionDrop,
					}),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSecurityGroupNoPort(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayInstanceSecurityGroupConfigNoPort[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.base", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      "0.0.0.0/0",
						DestPortFrom: nil,
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceSecurityGroupRuleMatch(name string, index int, expected *instance.SecurityGroupRule) resource.TestCheckFunc {
	return testAccCheckScalewayInstanceSecurityGroupRuleIs(name, expected.Direction, index, func(actual *instance.SecurityGroupRule) error {
		if !securityGroupRuleEquals(expected, actual) {
			return fmt.Errorf("security group does not match %v, %v", actual, expected)
		}
		return nil
	})
}

func testAccCheckScalewayInstanceSecurityGroupRuleIs(name string, direction instance.SecurityGroupRuleDirection, index int, test func(rule *instance.SecurityGroupRule) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		resRules, err := instanceApi.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
			SecurityGroupID: ID,
			Zone:            zone,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}
		sort.Slice(resRules.Rules, func(i, j int) bool {
			return resRules.Rules[i].Position < resRules.Rules[j].Position
		})
		apiRules := map[instance.SecurityGroupRuleDirection][]*instance.SecurityGroupRule{
			instance.SecurityGroupRuleDirectionInbound:  {},
			instance.SecurityGroupRuleDirectionOutbound: {},
		}

		for _, apiRule := range resRules.Rules {
			if apiRule.Editable == false {
				continue
			}
			apiRules[apiRule.Direction] = append(apiRules[apiRule.Direction], apiRule)
		}

		return test(apiRules[direction][index])
	}
}

func testAccCheckScalewayInstanceSecurityGroupExists(n string) resource.TestCheckFunc {
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

func testAccCheckScalewayInstanceSecurityGroupDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta().(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_instance_security_group" {
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

func testSweepComputeInstanceSecurityGroup(region string) error {
	scwClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	instanceAPI := instance.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the security groups in (%s)", region)

	listResp, err := instanceAPI.ListSecurityGroups(&instance.ListSecurityGroupsRequest{})
	if err != nil {
		return fmt.Errorf("error listing security groups in sweeper: %s", err)
	}

	for _, securityGroup := range listResp.SecurityGroups {
		// Can't delete default security group.
		if securityGroup.OrganizationDefault {
			continue
		}
		err = instanceAPI.DeleteSecurityGroup(&instance.DeleteSecurityGroupRequest{
			SecurityGroupID: securityGroup.ID,
		})
		if err != nil {
			return fmt.Errorf("error deleting security groups in sweeper: %s", err)
		}
	}

	return nil
}
