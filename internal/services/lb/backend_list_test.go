package lb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	lbtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb/testfuncs"
)

func TestAccListLbBackends_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListLbBackends_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             lbtestfuncs.IsIPDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb_ip" "ip1" {
					  zone = "fr-par-1"
					}

					resource "scaleway_lb" "lb1" {
					  ip_ids = [scaleway_lb_ip.ip1.id]
					  zone   = "fr-par-1"
					  name   = "test-acc-backend-list"
					  type   = "LB-S"
					}

					resource "scaleway_lb_backend" "b1" {
					  lb_id            = scaleway_lb.lb1.id
					  name             = "test-backend-list-one"
					  forward_protocol = "http"
					  forward_port     = 80
					}

					resource "scaleway_lb_backend" "b2" {
					  lb_id            = scaleway_lb.lb1.id
					  name             = "test-backend-list-two"
					  forward_protocol = "tcp"
					  forward_port     = 443
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_lb_backend" "by_lb" {
					  provider = scaleway

					  config {
					    zones  = ["fr-par-1"]
					    lb_ids = [scaleway_lb.lb1.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_lb_backend.by_lb", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_lb_backend" "by_name" {
					  provider = scaleway

					  config {
					    zones  = ["fr-par-1"]
					    lb_ids = [scaleway_lb.lb1.id]
					    name   = "test-backend-list-one"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_lb_backend.by_name", 1),
				},
			},
		},
	})
}
