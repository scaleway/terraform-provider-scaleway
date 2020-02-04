package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"testing"
)

func TestAccScalewayInstanceSecurityGroupRules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_instance_security_group sg01 {
						external_rules = true
					}
					resource scaleway_instance_security_group_rules sgrs01 {
						security_group_id = scaleway_instance_security_group.sg01.id
					}
				`,
			},
			{
				Config: `
					resource scaleway_instance_security_group sg01 {
						inbound_rule {
							action = "accept"
							port = 80
							ip_range = "0.0.0.0/0"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("scaleway_instance_security_group.sg01"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.sg01", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      expandIPNet("0.0.0.0/0"),
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
				),
			},
			{
				Config: `
					resource scaleway_instance_security_group sg01 {
						external_rules = true
					}
					resource scaleway_instance_security_group_rules sgrs01 {
						security_group_id = scaleway_instance_security_group.sg01.id
					}
				`,
			},
			{
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
					}
				`,
				/*Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("scaleway_instance_security_group.sg01"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.sg01", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      expandIPNet("0.0.0.0/0"),
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
					resource.TestCheckResourceAttrPair("scaleway_instance_security_group.sg01", "id", "scaleway_instance_security_group_rules.sgrs01", "security_group_id"),
				),*/
			},
			/*{
				ResourceName:      "scaleway_instance_security_group.sgrs01",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
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
						outbound_rule {
							action = "accept"
							port = 443
							ip_range = "127.0.0.1/0"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("scaleway_instance_security_group.sg01"),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.sg01", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionInbound,
						IPRange:      expandIPNet("0.0.0.0/0"),
						DestPortFrom: scw.Uint32Ptr(80),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
					testAccCheckScalewayInstanceSecurityGroupRuleMatch("scaleway_instance_security_group.sg01", 0, &instance.SecurityGroupRule{
						Direction:    instance.SecurityGroupRuleDirectionOutbound,
						IPRange:      expandIPNet("0.0.0.0/0"),
						DestPortFrom: scw.Uint32Ptr(443),
						DestPortTo:   nil,
						Protocol:     instance.SecurityGroupRuleProtocolTCP,
						Action:       instance.SecurityGroupRuleActionAccept,
					}),
					resource.TestCheckResourceAttrPair("scaleway_instance_security_group.sg01", "id", "scaleway_instance_security_group_rules.sgrs01", "security_group_id"),
				),
			},
			{
				Config: `resource scaleway_instance_security_group sg01 {
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists("scaleway_instance_security_group.sg01"),
				),
			},*/
		},
	})
}
