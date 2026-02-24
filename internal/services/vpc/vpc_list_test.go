package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListVPCs_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Query: true,
				Config: `
				list "scaleway_vpc" "fr-par" {
					config {
						region = "fr-par"
 					}
				}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("scaleway_vpc.fr-par", map[string]knownvalue.Check{
						"region": knownvalue.StringExact("fr-par"),
					}),
				},
			},
		},
	})
}
