package rdb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
	rdbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb/testfuncs"
)

func TestAccListRDBInstances_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListRDBInstances_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := rdbchecks.GetLatestEngineVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "main" {}

					resource "scaleway_rdb_instance" "main" {
					  project_id      = scaleway_account_project.main.id
					  name            = "test-rdb-list-1"
					  node_type       = "db-dev-s"
					  engine          = %q
					  is_ha_cluster   = false
					  disable_backup  = true
					  user_name       = "my_initial_user"
					  password        = "thiZ_is_v&ry_s3cret"
					  tags            = ["list-test"]
					}
				`, latestEngineVersion),
			},
			{
				Query: true,
				Config: `
					list "scaleway_rdb_instance" "all" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_instance.all", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_rdb_instance" "by_name" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    name        = "test-rdb-list-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_instance.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_rdb_instance" "by_tag" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    tags        = ["list-test"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_instance.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_rdb_instance" "without_maintenances" {
					  provider = scaleway

					  config {
					    regions           = ["fr-par"]
					    project_ids       = [scaleway_account_project.main.id]
					    has_maintenances  = false
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_instance.without_maintenances", 1),
				},
			},
		},
	})
}
