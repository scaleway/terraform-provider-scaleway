package rdb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	rdbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb/testfuncs"
)

func TestAccListRDBSnapshots_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListRDBSnapshots_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectID := rdbchecks.ListProjectID(tt)
	latestEngineVersion := rdbchecks.GetLatestEngineVersion(tt, postgreSQLEngineName)
	snapshotName := "tf-test-rdb-snapshot-list"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			rdbchecks.IsInstanceDestroyed(tt),
			IsSnapshotDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
					  project_id         = %q
					  name               = "tf-test-rdb-snapshot-list-inst"
					  node_type          = "db-dev-s"
					  engine             = %q
					  is_ha_cluster      = false
					  disable_backup     = true
					  user_name          = "my_initial_user"
					  password           = "thiZ_is_v&ry_s3cret"
					  volume_type        = "sbs_5k"
					  volume_size_in_gb  = 10
					}

					resource "scaleway_rdb_snapshot" "main" {
					  name        = "%s"
					  instance_id = scaleway_rdb_instance.main.id
					  depends_on  = [scaleway_rdb_instance.main]
					}
				`, projectID, latestEngineVersion, snapshotName),
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_rdb_snapshot" "by_instance" {
					  provider = scaleway

					  config {
					    regions      = ["fr-par"]
					    project_ids  = [%q]
					    instance_ids = [scaleway_rdb_instance.main.id]
					    name         = "tf-test-rdb-snapshot-list"
					  }
					}
				`, projectID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_snapshot.by_instance", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_rdb_snapshot" "wildcard" {
					  provider = scaleway

					  config {
					    regions      = ["fr-par"]
					    project_ids  = [%q]
					    instance_ids = ["*"]
					    name         = "tf-test-rdb-snapshot-list"
					  }
					}
				`, projectID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_rdb_snapshot.wildcard", 1),
				},
			},
		},
	})
}
