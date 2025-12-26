package vpcgw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
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
						name = "tf-test-vpcgw-action-refresh-ssh-keys"
						type = "VPC-GW-S"
						bastion_enabled = true

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_vpc_public_gateway_refresh_ssh_keys_action.main]
							}
						}
					}

					action "scaleway_vpc_public_gateway_refresh_ssh_keys_action" "main" {
						config {
							gateway_id = scaleway_vpc_public_gateway.main.id
							wait = true
						}
					}

					data "scaleway_vpc_public_gateway" "main" {
						public_gateway_id = scaleway_vpc_public_gateway.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_vpc_public_gateway.main", "name", "tf-test-vpcgw-action-refresh-ssh-keys"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_public_gateway.main", "status", "running"),
				),
			},
		},
	})
}
