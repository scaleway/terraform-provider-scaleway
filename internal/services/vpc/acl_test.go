package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpcSDK "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc"
)

func TestAccACL_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isACLDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-vpc-acl-basic"
					}
					
					resource "scaleway_vpc_acl" "acl01" {
					  vpc_id   = scaleway_vpc.vpc01.id
					  is_ipv6  = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isACLPresent(tt, "scaleway_vpc_acl.acl01"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_acl.acl01", "vpc_id", "scaleway_vpc.vpc01", "id"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "is_ipv6", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "default_policy", "accept"),
				),
			},
		},
	})
}

func TestAccACL_WithRules(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isACLDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-vpc-acl"
					}
					
					resource "scaleway_vpc_acl" "acl01" {
					  vpc_id   = scaleway_vpc.vpc01.id
					  is_ipv6  = false
					  rules {
						protocol      = "TCP"
						src_port_low  = 0
						src_port_high = 0
						dst_port_low  = 80
						dst_port_high = 80
						source        = "0.0.0.0/0"
						destination   = "0.0.0.0/0"
						description   = "Allow HTTP traffic from any source"
						action        = "accept"
					  }
					  default_policy = "drop"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isACLPresent(tt, "scaleway_vpc_acl.acl01"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_acl.acl01", "vpc_id", "scaleway_vpc.vpc01", "id"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "is_ipv6", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "default_policy", "drop"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.#", "1"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.src_port_low", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.src_port_high", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.dst_port_low", "80"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.dst_port_high", "80"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.source", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.destination", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.description", "Allow HTTP traffic from any source"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.action", "accept"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-vpc-acl"
					}
					
					resource "scaleway_vpc_acl" "acl01" {
					  vpc_id   = scaleway_vpc.vpc01.id
					  is_ipv6  = false
					  rules {
						protocol      = "TCP"
						src_port_low  = 0
						src_port_high = 0
						dst_port_low  = 80
						dst_port_high = 80
						source        = "0.0.0.0/0"
						destination   = "0.0.0.0/0"
						description   = "Allow HTTP traffic from any source"
						action        = "accept"
					  }
					  rules {
						protocol      = "TCP"
						src_port_low  = 0
						src_port_high = 0
						dst_port_low  = 443
						dst_port_high = 443
						source        = "0.0.0.0/0"
						destination   = "0.0.0.0/0"
						description   = "Allow HTTPS traffic from any source"
						action        = "accept"
					  }
					  default_policy = "drop"
					}
					
				`,
				Check: resource.ComposeTestCheckFunc(
					isACLPresent(tt, "scaleway_vpc_acl.acl01"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_acl.acl01", "vpc_id", "scaleway_vpc.vpc01", "id"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "is_ipv6", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "default_policy", "drop"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.#", "2"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.src_port_low", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.src_port_high", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.dst_port_low", "80"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.dst_port_high", "80"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.source", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.destination", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.description", "Allow HTTP traffic from any source"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.0.action", "accept"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.src_port_low", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.src_port_high", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.dst_port_low", "443"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.dst_port_high", "443"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.source", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.destination", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.description", "Allow HTTPS traffic from any source"),
					resource.TestCheckResourceAttr("scaleway_vpc_acl.acl01", "rules.1.action", "accept"),
				),
			},
		},
	})
}

func isACLPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, region, ID, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetACL(&vpcSDK.GetACLRequest{
			VpcID:  ID,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isACLDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_acl" {
				continue
			}

			vpcAPI, region, ID, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcAPI.GetACL(&vpcSDK.GetACLRequest{
				VpcID:  ID,
				Region: region,
			})

			if err == nil {
				return fmt.Errorf("acl (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
