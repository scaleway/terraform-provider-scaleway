package instance_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/stretchr/testify/require"
)

func TestAccSecurityGroupRules_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSecurityGroupDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// Simple empty configuration
				Config: `
						resource scaleway_instance_security_group sg01 {
							external_rules = true
						}

						resource scaleway_instance_security_group_rules sgrs01 {
							security_group_id = scaleway_instance_security_group.sg01.id
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.#", "0"),
				),
			},
			{
				// We test that we can add some rules, and they stay in correct orders
				Config: `
							resource scaleway_instance_security_group sg01 {
								external_rules = true
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
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg01"),
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
			{
				ResourceName:      "scaleway_instance_security_group_rules.sgrs01",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// We test that we can remove some rules
				Config: `
						resource scaleway_instance_security_group sg01 {
							external_rules = true
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
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg01"),
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
				ResourceName:      "scaleway_instance_security_group_rules.sgrs01",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// We test that we can remove all rules
				Config: `
						resource scaleway_instance_security_group sg01 {
							external_rules = true
						}
				
						resource scaleway_instance_security_group_rules sgrs01 {
							security_group_id = scaleway_instance_security_group.sg01.id
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.#", "0"),
				),
			},
			{
				ResourceName:      "scaleway_instance_security_group_rules.sgrs01",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSecurityGroupRules_IPRanges(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	config := `
			resource scaleway_instance_security_group sg01 {
				external_rules = true
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSecurityGroupDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ImportState:  true,
				ResourceName: "scaleway_instance_security_group_rules.sgrs01",
				Config:       config,
				Check: resource.ComposeTestCheckFunc(
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "6"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.ip_range", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.ip_range", "1.2.0.0/16"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.0.ip_range", "1.2.3.0/32"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "2002::/24"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "2002:0:0:1234::/64"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "outbound_rule.1.ip_range", "2002::1234:abcd:ffff:c0a8:101/128"),
				),
			},
			{
				ResourceName:            "scaleway_instance_security_group_rules.sgrs01",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"outbound_rule.0.ip"},
			},
		},
	})
}

func TestAccSecurityGroupRules_ReapplyConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	config := `
			resource scaleway_instance_security_group sg01 {
				external_rules = true
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSecurityGroupDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ImportState:  true,
				ResourceName: "scaleway_instance_security_group_rules.sgrs01",
				Config:       config,
				Check: resource.ComposeTestCheckFunc(
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg01"),
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
			{
				ResourceName:      "scaleway_instance_security_group_rules.sgrs01",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSecurityGroupRules_Priority(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ipnetZero, err := types.ExpandIPNet("0.0.0.0/0")
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSecurityGroupDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_instance_security_group sg01 {
						external_rules = true
					}

					resource scaleway_instance_security_group_rules sgrs01 {
						security_group_id = scaleway_instance_security_group.sg01.id
						inbound_rule {
							action = "accept"
							port = 8080
							ip_range = "0.0.0.0/0"
							priority = 10
						}
						inbound_rule {
							action = "drop"
							port = 80
							ip_range = "0.0.0.0/0"
							priority = 20
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isSecurityGroupPresent(tt, "scaleway_instance_security_group.sg01"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.port", "8080"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.0.priority", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.action", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.port", "80"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group_rules.sgrs01", "inbound_rule.1.priority", "20"),
					// Verify that the API stores the rules in the priority order.
					isSecurityGroupRuleMatching(tt, "scaleway_instance_security_group_rules.sgrs01", 0, &instanceSDK.SecurityGroupRule{
						Direction:    instanceSDK.SecurityGroupRuleDirectionInbound,
						IPRange:      ipnetZero,
						DestPortFrom: new(uint32(8080)),
						DestPortTo:   nil,
						Protocol:     instanceSDK.SecurityGroupRuleProtocolTCP,
						Action:       instanceSDK.SecurityGroupRuleActionAccept,
					}),
					isSecurityGroupRuleMatching(tt, "scaleway_instance_security_group_rules.sgrs01", 1, &instanceSDK.SecurityGroupRule{
						Direction:    instanceSDK.SecurityGroupRuleDirectionInbound,
						IPRange:      ipnetZero,
						DestPortFrom: new(uint32(80)),
						DestPortTo:   nil,
						Protocol:     instanceSDK.SecurityGroupRuleProtocolTCP,
						Action:       instanceSDK.SecurityGroupRuleActionDrop,
					}),
				),
			},
		},
	})
}
