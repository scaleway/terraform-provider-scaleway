package s2svpn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceCustomerGateway_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "customer_ip" {}

					resource "scaleway_s2s_vpn_customer_gateway" "main" {
						name        = "tf-test-customer-gateway-ds"
						asn         = 65000
						ipv4_public = scaleway_instance_ip.customer_ip.address
					}
				`,
			},
			{
				Config: `
					resource "scaleway_instance_ip" "customer_ip" {}

					resource "scaleway_s2s_vpn_customer_gateway" "main" {
						name        = "tf-test-customer-gateway-ds"
						asn         = 65000
						ipv4_public = scaleway_instance_ip.customer_ip.address
					}

					data "scaleway_s2s_vpn_customer_gateway" "by_name" {
						name = scaleway_s2s_vpn_customer_gateway.main.name
					}

					data "scaleway_s2s_vpn_customer_gateway" "by_id" {
						customer_gateway_id = scaleway_s2s_vpn_customer_gateway.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(tt, "scaleway_s2s_vpn_customer_gateway.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_customer_gateway.by_name", "name",
						"scaleway_s2s_vpn_customer_gateway.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_customer_gateway.by_name", "asn",
						"scaleway_s2s_vpn_customer_gateway.main", "asn"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_customer_gateway.by_id", "customer_gateway_id",
						"scaleway_s2s_vpn_customer_gateway.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_customer_gateway.by_id", "name",
						"scaleway_s2s_vpn_customer_gateway.main", "name"),
				),
			},
		},
	})
}
