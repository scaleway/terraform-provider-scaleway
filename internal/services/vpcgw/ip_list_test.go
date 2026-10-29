package vpcgw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpcgwchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw/testfuncs"
)

func TestAccListVPCPublicGatewayIPs_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListVPCPublicGatewayIPs_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpcgwchecks.IsIPDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_public_gateway_ip" "ip1" {
					  zone = "fr-par-1"
					  tags = ["tf-acc-pgw-ip-list"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_public_gateway_ip" "by_project" {
					  provider = scaleway

					  config {
					    zones       = ["fr-par-1"]
					    project_ids = [scaleway_vpc_public_gateway_ip.ip1.project_id]
					    tags        = ["tf-acc-pgw-ip-list"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_vpc_public_gateway_ip.by_project", 1),
				},
			},
		},
	})
}
