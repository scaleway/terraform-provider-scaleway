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

func TestAccListRDBDatabases_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListRDBDatabases_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := rdbchecks.GetLatestEngineVersion(tt, postgreSQLEngineName)
	instanceName := "tf-test-rdb-db-list-4823616331525728375"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "main" {}

					resource "scaleway_rdb_instance" "main" {
					  project_id    = scaleway_account_project.main.id
					  name          = "%s"
					  node_type     = "db-dev-s"
					  engine        = %q
					  is_ha_cluster = false
					  tags          = ["rdb-db-list-test"]
					}

					resource "scaleway_rdb_database" "main" {
					  instance_id = scaleway_rdb_instance.main.id
					  name        = "tfdb_list_2078923807392667811"
					}
				`, instanceName, latestEngineVersion),
			},
			{
				Query: true,
				Config: `
					list "scaleway_rdb_database" "by_instance" {
					  provider = scaleway

					  config {
					    regions       = ["fr-par"]
					    project_ids   = [scaleway_account_project.main.id]
					    instance_ids  = [scaleway_rdb_instance.main.id]
					    name          = "tfdb_list_2078923807392667811"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_database.by_instance", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_rdb_database" "wildcard" {
					  provider = scaleway

					  config {
					    regions      = ["fr-par"]
					    project_ids  = [scaleway_account_project.main.id]
					    instance_ids = ["*"]
					    name         = "tfdb_list_2078923807392667811"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_database.wildcard", 1),
				},
			},
		},
	})
}
