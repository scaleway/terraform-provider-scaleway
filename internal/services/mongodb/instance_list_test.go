package mongodb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListMongoDBInstances_Basic(t *testing.T) {

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "main" {}

					resource "scaleway_mongodb_instance" "main" {
					  project_id  = scaleway_account_project.main.id
					  name        = "test-mongodb-list-1"
					  version     = "7.0"
					  node_type   = "MGDB-PLAY2-NANO"
					  node_number = 1
					  user_name   = "my_initial_user"
					  password    = "thiZ_is_v&ry_s3cret"
					  tags        = ["list-test"]
					}

					resource "scaleway_mongodb_instance" "alt" {
					  project_id  = scaleway_account_project.main.id
					  name        = "test-mongodb-list-2"
					  version     = "7.0"
					  node_type   = "MGDB-PLAY2-NANO"
					  node_number = 1
					  user_name   = "my_initial_user"
					  password    = "thiZ_is_v&ry_s3cret"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_mongodb_instance" "all" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_mongodb_instance.all", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_mongodb_instance" "by_name" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    name        = "test-mongodb-list-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_mongodb_instance.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_mongodb_instance" "by_tag" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    tags        = ["list-test"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_mongodb_instance.by_tag", 1),
				},
			},
		},
	})
}
