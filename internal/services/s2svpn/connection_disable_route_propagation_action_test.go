package s2svpn_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/s2svpn"
)

func TestAccActionS2SVPNConnectionDisableRoutePropagation_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionS2SVPNConnectionDisableRoutePropagation_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckConnectionDestroy(tt),
			testAccCheckVPNGatewayDestroy(tt),
			testAccCheckCustomerGatewayDestroy(tt),
			testAccCheckRoutingPolicyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-test-vpc-disable-route-prop"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						ipv4_subnet {
							subnet = "10.0.0.0/24"
						}
					}

					resource "scaleway_instance_ip" "customer_ip" {}

					resource "scaleway_s2s_vpn_gateway" "main" {
						name               = "tf-test-vpn-gateway-disable-route-prop"
						gateway_type       = "VGW-S"
						private_network_id = scaleway_vpc_private_network.main.id
						region             = "fr-par"
						zone               = "fr-par-1"
					}

					resource "scaleway_s2s_vpn_customer_gateway" "main" {
						name        = "tf-test-customer-gateway-disable-route-prop"
						ipv4_public = scaleway_instance_ip.customer_ip.address
						asn         = 65000
						region      = "fr-par"
					}

					resource "scaleway_s2s_vpn_routing_policy" "main" {
						name              = "tf-test-routing-policy-disable-route-prop"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.0.0/24"]
						region            = "fr-par"
					}

					resource "scaleway_s2s_vpn_connection" "main" {
						name                     = "tf-test-connection-disable-route-prop"
						vpn_gateway_id            = scaleway_s2s_vpn_gateway.main.id
						customer_gateway_id       = scaleway_s2s_vpn_customer_gateway.main.id
						initiation_policy         = "customer_gateway"
						enable_route_propagation  = true
						region                    = "fr-par"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_s2s_vpn_connection_disable_route_propagation.main]
							}
						}

						bgp_config_ipv4 {
							routing_policy_id = scaleway_s2s_vpn_routing_policy.main.id
							private_ip        = "169.254.0.1/30"
							peer_private_ip   = "169.254.0.2/30"
						}

						ikev2_ciphers {
							encryption = "aes256"
							integrity  = "sha256"
							dh_group   = "modp2048"
						}

						esp_ciphers {
							encryption = "aes256"
							integrity  = "sha256"
							dh_group   = "modp2048"
						}
					}

					action "scaleway_s2s_vpn_connection_disable_route_propagation" "main" {
						config {
							connection_id = scaleway_s2s_vpn_connection.main.id
							region        = "fr-par"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(tt, "scaleway_s2s_vpn_connection.main"),
					isRoutePropagationDisabled(tt, "scaleway_s2s_vpn_connection.main"),
				),
			},
		},
	})
}

func isRoutePropagationDisabled(tt *acctest.TestTools, connectionResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[connectionResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", connectionResourceName)
		}

		api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to get API and parse connection ID: %w", err)
		}

		connection, err := api.GetConnection(&s2s_vpn.GetConnectionRequest{
			Region:       region,
			ConnectionID: id,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to get connection: %w", err)
		}

		if connection == nil {
			return fmt.Errorf("connection %s not found", id)
		}

		if connection.RoutePropagationEnabled {
			return fmt.Errorf("connection %s has route propagation enabled, expected disabled", id)
		}

		return nil
	}
}
