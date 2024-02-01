package scaleway

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScalewayIPAMIPReverseDNS_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	testDNSZone := "tf-reverse-ipam." + testDomain
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_instance_ip" "ip01" {
				  type = "routed_ipv6"
				}
				
				resource "scaleway_instance_server" "srv01" {
				  name   = "tf-tests-instance-server-ips"
				  ip_ids = [scaleway_instance_ip.ip01.id]
				  image  = "ubuntu_jammy"
				  type   = "PRO2-XXS"
				  state  = "stopped"
				}
				
				data "scaleway_ipam_ip" "ipam01" {
				  resource {
					id   = scaleway_instance_server.srv01.id
					type = "instance_server"
				  }
				  type = "ipv6"
				}
				`,
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_instance_ip" "ip01" {
				  type = "routed_ipv6"
				}

				resource "scaleway_instance_server" "srv01" {
				  name   = "tf-tests-instance-server-ips"
				  ip_ids = [scaleway_instance_ip.ip01.id]
				  image  = "ubuntu_jammy"
				  type   = "PRO2-XXS"
				  state  = "stopped"
				}

				data "scaleway_ipam_ip" "ipam01" {
				  resource {
					id   = scaleway_instance_server.srv01.id
					type = "instance_server"
				  }
				  type = "ipv6"
				}

				resource "scaleway_domain_record" "tf_AAAA" {
				  dns_zone = %[1]q
				  name     = ""
				  type     = "AAAA"
				  data     = cidrhost(data.scaleway_ipam_ip.ipam01.address_cidr, 42)
				  ttl      = 3600
				  priority = 1
				}

				resource "scaleway_ipam_ip_reverse_dns" "base" {
				  ipam_ip_id = data.scaleway_ipam_ip.ipam01.id

                  hostname   = %[1]q
				  address    = cidrhost(data.scaleway_ipam_ip.ipam01.address_cidr, 42)
				}

				output "calculated_ip_address" {
				  value = cidrhost(data.scaleway_ipam_ip.ipam01.address_cidr, 42)
				}
				`, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrExpectedIPAddress("scaleway_ipam_ip_reverse_dns.base"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip_reverse_dns.base", "hostname", testDNSZone),
				),
			},
		},
	})
}

func testCheckResourceAttrExpectedIPAddress(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()

		if output, ok := rootModule.Outputs["calculated_ip_address"]; ok && output != nil {
			expectedIPAddress := output.Value.(string)
			return resource.TestCheckResourceAttr(resourceName, "address", expectedIPAddress)(s)
		}
		return errors.New("calculated_ip_address output not set or is nil")
	}
}
