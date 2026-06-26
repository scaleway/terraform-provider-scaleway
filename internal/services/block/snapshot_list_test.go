package block_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccListBlockSnapshots_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListBlockSnapshots_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "vol1" {
						name       = "test-vol-snapshot-list-1"
						size_in_gb = 10
						iops       = 5000
					}

					resource "scaleway_block_volume" "vol2" {
						name       = "test-vol-snapshot-list-2"
						size_in_gb = 10
						iops       = 5000
					}

					resource "scaleway_block_snapshot" "snap1" {
						name      = "test-snapshot-list-1"
						volume_id = scaleway_block_volume.vol1.id
					}

					resource "scaleway_block_snapshot" "snap2" {
						name      = "test-snapshot-list-2"
						volume_id = scaleway_block_volume.vol2.id
						tags      = ["test-tag"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_block_snapshot" "all" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
							volume_ids  = ["*"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_block_snapshot.all", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_block_snapshot" "by_volume" {
						provider = scaleway

						config {
							zones       = [scaleway_block_volume.vol1.zone]
							project_ids = [scaleway_block_volume.vol1.project_id]
							volume_ids  = [scaleway_block_volume.vol1.id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_snapshot.by_volume", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_block_snapshot" "by_name" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
							volume_ids  = ["*"]
							name        = "test-snapshot-list-1"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_snapshot.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_block_snapshot" "by_tag" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
							volume_ids  = ["*"]
							tags        = ["test-tag"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_snapshot.by_tag", 1),
				},
			},
		},
	})
}
