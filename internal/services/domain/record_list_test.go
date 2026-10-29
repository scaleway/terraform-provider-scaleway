package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListDomainRecords_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListDomainRecords_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	testDNSZone := "test-record-list"
	projectID := testAccDomainZoneProjectID(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckDomainRecordDestroy(tt),
			testAccCheckDomainZoneDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain     = %q
						subdomain  = %q
						project_id = %q
					}

					resource "scaleway_domain_record" "test" {
						dns_zone = scaleway_domain_zone.test.id
						type     = "A"
						data     = "127.0.0.1"
						ttl      = 3600
					}
				`, acctest.TestDomain, testDNSZone, projectID),
			},
			{
				Query: true,
				Config: `
					list "scaleway_domain_record" "by_zone" {
						provider = scaleway

						config {
							project_ids = [scaleway_domain_record.test.project_id]
							dns_zones   = [scaleway_domain_record.test.dns_zone]
							type        = "A"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_domain_record.by_zone", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_domain_record" "wildcard_zones" {
						provider = scaleway

						config {
							project_ids = [scaleway_domain_record.test.project_id]
							dns_zones   = ["*"]
							type        = "A"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_domain_record.wildcard_zones", 1),
				},
			},
		},
	})
}
