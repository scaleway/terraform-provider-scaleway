package block_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccListBlockVolumes_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListBlockVolumes_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	testDefaultZone, _ := tt.Meta.ScwClient().GetDefaultZone()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "vol1" {
						name       = "test-volume-list-1"
						size_in_gb = 10
						iops       = 5000
					}
				`,
			},

			{
				Config: `
					resource "scaleway_block_volume" "vol1" {
						name       = "test-volume-list-1"
						size_in_gb = 10
						iops       = 5000
					}

					resource "scaleway_block_volume" "vol2" {
						name       = "test-volume-list-2"
						size_in_gb = 10
						iops       = 5000
					}
				`,
			},

			{
				Config: `
					resource "scaleway_block_volume" "vol1" {
						name       = "test-volume-list-1"
						size_in_gb = 10
						iops       = 5000
					}

					resource "scaleway_block_volume" "vol2" {
						name       = "test-volume-list-2"
						size_in_gb = 10
						iops       = 5000
					}

					resource "scaleway_block_volume" "vol3" {
						name       = "test-volume-list-3"
						size_in_gb = 10
						iops       = 5000
						zone       = "pl-waw-2"
					}
				`,
			},

			{
				Query: true,
				Config: `
					list "scaleway_block_volume" "all" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_volume.all", 3),
				},
			},

			{
				Query: true,
				Config: `
					list "scaleway_block_volume" "by_zone_pl_waw_2" {
						provider = scaleway

						config {
							project_ids = ["*"]
							zones       = ["pl-waw-2"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_volume.by_zone_pl_waw_2", 1),
				},
			},

			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_block_volume" "by_zone_default" {
						provider = scaleway

						config {
							project_ids = ["*"]
							zones       = ["%s"]
						}
					}
				`, testDefaultZone),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_block_volume.by_zone_default", 2),
				},
			},

			{
				Query: true,
				Config: `
					list "scaleway_block_volume" "by_name" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
							name        = "test-volume-list-1"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_volume.by_name", 1),
				},
			},
		},
	})
}

func TestAccListBlockVolumes_WithTags(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListBlockVolumes_WithTags because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "vol1" {
						name       = "test-volume-tags-1"
						size_in_gb = 10
						iops       = 5000
						tags       = ["test-tag", "common"]
					}

					resource "scaleway_block_volume" "vol2" {
						name       = "test-volume-tags-2"
						size_in_gb = 10
						iops       = 5000
						tags       = ["test-tag"]
					}

					resource "scaleway_block_volume" "vol3" {
						name       = "test-volume-tags-3"
						size_in_gb = 10
						iops       = 5000
						tags       = ["other-tag"]
					}
				`,
			},

			{
				Query: true,
				Config: `
					list "scaleway_block_volume" "by_tag" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
							tags        = ["test-tag"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_volume.by_tag", 2),
				},
			},

			{
				Query: true,
				Config: `
					list "scaleway_block_volume" "by_other_tag" {
						provider = scaleway

						config {
							zones       = ["*"]
							project_ids = ["*"]
							tags        = ["other-tag"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_block_volume.by_other_tag", 1),
				},
			},
		},
	})
}
