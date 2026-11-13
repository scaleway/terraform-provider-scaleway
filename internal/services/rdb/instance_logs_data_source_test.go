package rdb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	rdbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb/testfuncs"
)

func TestAccDataSourceInstanceLogs_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := rdbchecks.GetLatestEngineVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             rdbchecks.IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						name = "tf-tests-rdb-instance-logs"
						engine = %q
						node_type = "db-dev-s"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
					}

					data "scaleway_rdb_instance_logs" "main" {
						instance_id = scaleway_rdb_instance.main.id
						order_by = "created_at_desc"
					}
				`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					isInstancePresent(tt, "scaleway_rdb_instance.main"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_logs.main", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_logs.main", "region"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_logs.main", "instance_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_logs.main", "instance_logs.#"),
				),
			},
		},
	})
}
