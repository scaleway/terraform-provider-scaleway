package file_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListFileSystems_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListFSs_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_file_filesystem" "fs1" {
					name = "test-fs-01"
					size_in_gb = 100
					tags = ["foo", "bar"]
				}

				resource "scaleway_file_filesystem" "fs2" {
					name = "test-fs-02"
					size_in_gb = 200
				}`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_file_filesystem" "by_name" {
					  provider = scaleway
					
					  config {
						project_ids = ["*"]
						regions = ["fr-par"]
						name = "test-fs"
					  }
					}`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_file_filesystem.by_name", 2),
				},
			},
		},
	})
}
