package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
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
				// Create VPCs in different regions with different tags
				ConfigDirectory: config.StaticDirectory("testdata/vpc_list_basic/"),
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/vpc_list_basic/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					// Check that we can list all VPCs
					querycheck.ExpectLength("list.scaleway_vpc.all", 2),
					// Check that we can filter by region and tag
					querycheck.ExpectLength("list.scaleway_vpc.fr-par", 1),
					// Check that we can filter by name pattern
					querycheck.ExpectLength("list.scaleway_vpc.by_name", 2),
					// Verify specific VPC attributes in the results
					querycheck.ExpectIdentity("list.scaleway_vpc.fr-par", map[string]knownvalue.Check{
						"region": knownvalue.StringExact("fr-par"),
					}),
				},
			},
		},
	})
}
