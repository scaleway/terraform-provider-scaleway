package scaleway

import (
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestAccScalewayInstanceSecurityGroupRules_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	ipNetZero, err := expandIPNet("0.0.0.0/0")
	assert.NoError(t, err)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceSecurityGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				// Simple empty configuration
				Config: `
						resource scaleway_instance_security_group sg01 {
						}

						resource scaleway_instance_security_group_rules sgrs01 {
							security_group_id = scaleway_instance_security_group.sg01.id
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.#", "0"),
				),
			},
			{
				// We test that we can add some rules, and they stay in correct orders
				Config: `
							resource scaleway_instance_security_group sg01 {
							}

							resource scaleway_instance_security_group_rules sgrs01 {
								security_group_id = scaleway_instance_security_group.sg01.id
								inbound_rule {
									action = "accept"
									port = 80
									ip_range = "0.0.0.0/0"
								}
								inbound_rule {
									action = "drop"
									port = 443
									ip_range = "0.0.0.0/0"
								}
								outbound_rule {
									action = "accept"
									port = 80
									ip_range = "0.0.0.0/0"
								}
								outbound_rule {
									action = "drop"
									port = 443
									ip_range = "0.0.0.0/0"
								}
							}
						`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.port", "443"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.port", "443"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "0.0.0.0/0"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch(tt, "scaleway_instance_security_group_rules.sgrs01", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      ipNetZero,
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
				),
			},
			{
				// We test that we can remove some rules
				Config: `
						resource scaleway_instance_security_group sg01 {
						}
				
						resource scaleway_instance_security_group_rules sgrs01 {
							security_group_id = scaleway_instance_security_group.sg01.id
								inbound_rule {
									action = "drop"
									port = 443
									ip_range = "0.0.0.0/0"
								}
								outbound_rule {
									action = "drop"
									port = 443
									ip_range = "0.0.0.0/0"
								}
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.port", "443"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.port", "443"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.ip_range", "0.0.0.0/0"),
				),
			},
			{
				// We test that we can remove all rules
				Config: `
						resource scaleway_instance_security_group sg01 {
						}
				
						resource scaleway_instance_security_group_rules sgrs01 {
							security_group_id = scaleway_instance_security_group.sg01.id
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.#", "0"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSecurityGroupRules_IPRanges(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	config := `
			resource scaleway_instance_security_group sg01 {
			}

			resource scaleway_instance_security_group_rules sgrs01 {
				security_group_id = scaleway_instance_security_group.sg01.id
				inbound_rule {
					action = "accept"
					port = 80
					ip_range = "0.0.0.0/0"
				}
				inbound_rule {
					action = "drop"
					port = 443
					ip_range = "1.2.0.0/16"
				}
				outbound_rule {
					action = "accept"
					port = 80
					ip_range = "1.2.3.0/32"
				}
				outbound_rule {
					action = "drop"
					port = 443
					ip_range = "2002::/24"
				}
				outbound_rule {
					action = "drop"
					port = 443
					ip_range = "2002:0:0:1234::/64"
				}
				outbound_rule {
					action = "drop"
					port = 443
					ip_range = "2002::1234:abcd:ffff:c0a8:101/128"
				}

			}
		`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceSecurityGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ImportState:  true,
				ResourceName: "scaleway_instance_security_group_rules.sgrs01",
				Config:       config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "6"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.ip_range", "1.2.0.0/16"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.ip_range", "1.2.3.0/32"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "2002::/24"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "2002:0:0:1234::/64"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "2002::1234:abcd:ffff:c0a8:101/128"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSecurityGroupRules_Basic2(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	config := `
			resource scaleway_instance_security_group sg01 {
			}
			resource scaleway_instance_security_group_rules sgrs01 {
				security_group_id = scaleway_instance_security_group.sg01.id
				inbound_rule {
					action = "accept"
					port = 80
					ip_range = "0.0.0.0/0"
				}
				inbound_rule {
					action = "drop"
					port = 443
					ip_range = "0.0.0.0/0"
				}
				outbound_rule {
					action = "accept"
					port = 80
					ip_range = "0.0.0.0/0"
				}
				outbound_rule {
					action = "drop"
					port = 443
					ip_range = "0.0.0.0/0"
				}
			}
		`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceSecurityGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ImportState:  true,
				ResourceName: "scaleway_instance_security_group_rules.sgrs01",
				Config:       config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.port", "443"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.port", "443"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "0.0.0.0/0"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceSecurityGroupRuleMatch(tt *TestTools, name string, index int, expected *instance.SecurityGroupRule) resource.TestCheckFunc {
	return testAccCheckScalewayInstanceSecurityGroupRuleIs(tt, name, expected.Direction, index, func(actual *instance.SecurityGroupRule) error {
		if ok, _ := securityGroupRuleEquals(expected, actual); !ok {
			return fmt.Errorf("security group does not match %v, %v", actual, expected)
		}
		return nil
	})
}

func testAccCheckScalewayInstanceSecurityGroupRuleIs(tt *TestTools, name string, direction instance.SecurityGroupRuleDirection, index int, test func(rule *instance.SecurityGroupRule) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		resRules, err := instanceAPI.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
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
