package lb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListLBs_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListLBs_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isLbDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb" "lb1" {
					  zone = "fr-par-1"
					  name = "test-lb-list-1"
					  type = "LB-S"
					}

					resource "scaleway_lb" "lb2" {
					  zone = "fr-par-1"
					  name = "test-lb-list-2"
					  type = "LB-S"
					  tags = ["test-lb-list-tagged"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_lb" "by_name" {
					  provider = scaleway

					  config {
						zones       = ["fr-par-1"]
						project_ids = [scaleway_lb.lb1.project_id]
						name        = "test-lb-list-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_lb.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_lb" "by_tag" {
					  provider = scaleway

					  config {
						zones       = ["fr-par-1"]
						project_ids = [scaleway_lb.lb1.project_id]
						tags        = ["test-lb-list-tagged"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_lb.by_tag", 1),
				},
			},
		},
	})
}
