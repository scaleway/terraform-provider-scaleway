package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func TestAccScalewayVPCPublicGatewayIPReverseDns_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	testDNSZone := "tf-reverse-vpcgw." + testDomain
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_public_gateway_ip" "main" {}
					
					resource "scaleway_domain_record" "tf_A" {
						dns_zone = %[1]q
						name     = ""
						type     = "A"
                        data     = "${scaleway_vpc_public_gateway_ip.main.address}"
						ttl      = 3600
						priority = 1
					}

					resource "scaleway_vpc_public_gateway_ip_reverse_dns" "main" {
					    gateway_ip_id   = scaleway_vpc_public_gateway_ip.main.id
					    reverse         = %[1]q
						depends_on      = [scaleway_domain_record.tf_A]
					}
				`, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip_reverse_dns.main", "reverse", testDNSZone),
				),
			},
			{
				Config: `
					resource "scaleway_vpc_public_gateway_ip" "main" {}
				`,
				Check: testAccCheckScalewayVPCPublicGatewayIPDefaultReverse(tt, "scaleway_vpc_public_gateway_ip.main"),
			},
		},
	})
}

func testAccCheckScalewayVPCPublicGatewayIPDefaultReverse(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcgwAPI, zone, ID, err := scaleway.VpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		ip, err := vpcgwAPI.GetIP(&vpcgw.GetIPRequest{
			IPID: ID,
			Zone: zone,
		})
		if err != nil {
			return err
		}

		if *ip.Reverse != scaleway.FindDefaultReverse(ip.Address.String()) {
			return fmt.Errorf("reverse should be the same, %v is different than %v", *ip.Reverse, ip.Address.String())
		}

		return nil
	}
}
