package vpcgw_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	vpcgwSDK "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw"
)

func TestAccActionVPCGWRefreshSSHKeys_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionVPCGWRefreshSSHKeys_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_public_gateway" "main" {
						name = "test-vpcgw-action-refresh-ssh-keys"
						type = "VPC-GW-S"
						bastion_enabled = true

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_vpcgw_refresh_ssh_keys_action.main]
							}
						}
					}

					action "scaleway_vpcgw_refresh_ssh_keys_action" "main" {
						config {
							gateway_id = scaleway_vpc_public_gateway.main.id
							wait = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isGatewaySSHKeysRefreshed(tt, "scaleway_vpc_public_gateway.main"),
				),
			},
		},
	})
}

func isGatewaySSHKeysRefreshed(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		api, zone, ID, err := vpcgw.NewAPIWithZoneAndIDv2(tt.Meta, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse gateway ID: %w", err)
		}

		gateway, err := api.GetGateway(&vpcgwSDK.GetGatewayRequest{
			GatewayID: ID,
			Zone:      zone,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to get gateway: %w", err)
		}

		if gateway == nil {
			return fmt.Errorf("gateway %s not found", ID)
		}

		if gateway.Status != vpcgwSDK.GatewayStatusRunning {
			return fmt.Errorf("gateway %s is not running, status: %s", ID, gateway.Status)
		}

		return nil
	}
}
