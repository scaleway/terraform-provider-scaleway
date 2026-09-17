package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListDomainZones_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListDomainZones_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	testDNSZone := "test-zone"
	projectID := testAccDomainZoneProjectID(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckDomainZoneDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain     = %q
						subdomain  = %q
						project_id = %q
					}
				`, acctest.TestDomain, testDNSZone, projectID),
			},
			{
				Query: true,
				Config: `
					list "scaleway_domain_zone" "by_domain" {
						provider = scaleway

						config {
							project_ids = [scaleway_domain_zone.test.project_id]
							domains     = [scaleway_domain_zone.test.domain]
							dns_zones   = [scaleway_domain_zone.test.id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_domain_zone.by_domain", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_domain_zone" "wildcard_domains" {
						provider = scaleway

						config {
							project_ids = [scaleway_domain_zone.test.project_id]
							domains     = ["*"]
							dns_zones   = [scaleway_domain_zone.test.id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_domain_zone.wildcard_domains", 1),
				},
			},
		},
	})
}
