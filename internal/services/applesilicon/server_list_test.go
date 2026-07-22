package applesilicon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListAppleSiliconServers_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListAppleSiliconServers_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	testDefaultZone, _ := tt.Meta.ScwClient().GetDefaultZone()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "test-server-list-1"
						type = "M4-M"
						public_bandwidth = 1000000000
					}
				`,
			},
			{
				Query: true,
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "test-server-list-1"
						type = "M4-M"
						public_bandwidth = 1000000000
					}

					list "scaleway_apple_silicon_server" "all" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_apple_silicon_server.all", 1),
				},
			},
			{
				Query: true,
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "test-server-list-1"
						type = "M4-M"
						public_bandwidth = 1000000000
					}

					list "scaleway_apple_silicon_server" "by_zone_default" {
						provider = scaleway

						config {
							project_ids = ["*"]
							zones       = ["` + testDefaultZone.String() + `"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_apple_silicon_server.by_zone_default", 1),
				},
			},
		},
	})
}
