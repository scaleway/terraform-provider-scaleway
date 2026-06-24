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

func TestAccListRDBDatabaseBackups_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListRDBDatabaseBackups_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := rdbchecks.GetLatestEngineVersion(tt, postgreSQLEngineName)
	instanceName := "tf-test-rdb-backup-list-4823616331525728375"
	backupName := "tf_backup_list_2078923807392667811"

	var projectID, instanceID string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "main" {}
				`,
				Check: acctest.StoreResourceID("scaleway_account_project.main", &projectID),
			},
			{
				PreConfig: acctest.PreCheckWaitForRDBProjectIAM(tt, projectID),
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "main" {}

					resource "scaleway_rdb_instance" "main" {
					  project_id     = scaleway_account_project.main.id
					  name           = "%s"
					  node_type      = "db-dev-s"
					  engine         = %q
					  is_ha_cluster  = false
					  disable_backup = true
					  user_name      = "my_initial_user"
					  password       = "thiZ_is_v&ry_s3cret"
					  tags           = ["rdb-backup-list-test"]
					}

					resource "scaleway_rdb_database" "main" {
					  instance_id = scaleway_rdb_instance.main.id
					  name        = "tfdb_backup_list"
					}

					resource "scaleway_rdb_database_backup" "main" {
					  instance_id   = scaleway_rdb_instance.main.id
					  database_name = scaleway_rdb_database.main.name
					  name          = "%s"
					}
				`, instanceName, latestEngineVersion, backupName),
				Check: acctest.StoreResourceID("scaleway_rdb_instance.main", &instanceID),
			},
			{
				PreConfig: acctest.PreCheckWaitForRDBInstanceIAM(tt, instanceID),
				Query:     true,
				Config: fmt.Sprintf(`
					list "scaleway_rdb_database_backup" "by_instance" {
					  provider = scaleway

					  config {
					    regions      = ["fr-par"]
					    project_ids  = [scaleway_account_project.main.id]
					    instance_ids = [scaleway_rdb_instance.main.id]
					    name         = "%s"
					  }
					}
				`, backupName),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_database_backup.by_instance", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_rdb_database_backup" "wildcard" {
					  provider = scaleway

					  config {
					    regions      = ["fr-par"]
					    project_ids  = [scaleway_account_project.main.id]
					    instance_ids = ["*"]
					    name         = "%s"
					  }
					}
				`, backupName),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_database_backup.wildcard", 1),
				},
			},
		},
	})
}
