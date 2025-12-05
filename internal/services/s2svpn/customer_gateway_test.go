package s2svpn_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/s2svpn"
)

func TestAccCustomerGateway_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "main" {}

					resource "scaleway_s2s_vpn_customer_gateway" "main" {
						name        = "tf-test-customer-gateway"
						ipv4_public = scaleway_instance_ip.main.address
						asn         = 65000
						region      = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(tt, "scaleway_s2s_vpn_customer_gateway.main"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_customer_gateway.main", "name", "tf-test-customer-gateway"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_customer_gateway.main", "asn", "65000"),
					resource.TestCheckResourceAttrPair("scaleway_s2s_vpn_customer_gateway.main", "ipv4_public", "scaleway_instance_ip.main", "address"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_customer_gateway.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_customer_gateway.main", "created_at"),
				),
			},
		},
	})
}

func testAccCheckCustomerGatewayExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetCustomerGateway(&s2s_vpn.GetCustomerGatewayRequest{
			GatewayID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckCustomerGatewayDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_s2s_vpn_customer_gateway" {
				continue
			}

			api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetCustomerGateway(&s2s_vpn.GetCustomerGatewayRequest{
				GatewayID: id,
				Region:    region,
			})
			if err == nil {
				return fmt.Errorf("customer gateway (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
