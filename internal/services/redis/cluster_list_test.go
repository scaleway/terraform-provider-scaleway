package redis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	redisSDK "github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListRedisClusters_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListRedisClusters_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := getLatestVersion(tt)
	clusterName := sdkacctest.RandomWithPrefix("tf-test-redis-list")
	clusterTag := sdkacctest.RandomWithPrefix("redis-list-test")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "main" {}

					resource "scaleway_redis_cluster" "main" {
					  project_id   = scaleway_account_project.main.id
					  name         = "%s"
					  version      = "%s"
					  node_type    = "RED1-micro"
					  user_name    = "tf_test_user"
					  password     = "thiZ_is_v&ry_s3cret"
					  cluster_size = 1
					  zone         = "fr-par-1"
					  tags         = ["%s"]
					  depends_on   = [scaleway_account_project.main]
					}
				`, clusterName, latestVersion, clusterTag),
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_redis_cluster" "all" {
					  provider = scaleway

					  config {
					    zones       = ["fr-par-1"]
					    project_ids = [scaleway_account_project.main.id]
					    tags        = ["%s"]
					  }
					}
				`, clusterTag),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_redis_cluster.all", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_redis_cluster" "by_name" {
					  provider = scaleway

					  config {
					    zones       = ["fr-par-1"]
					    project_ids = [scaleway_account_project.main.id]
					    name        = "%s"
					  }
					}
				`, clusterName),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_redis_cluster.by_name", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_redis_cluster" "by_tag" {
					  provider = scaleway

					  config {
					    zones       = ["fr-par-1"]
					    project_ids = [scaleway_account_project.main.id]
					    tags        = ["%s"]
					  }
					}
				`, clusterTag),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_redis_cluster.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_redis_cluster" "by_version" {
					  provider = scaleway

					  config {
					    zones       = ["fr-par-1"]
					    project_ids = [scaleway_account_project.main.id]
					    name        = "%s"
					    version     = "%s"
					  }
					}
				`, clusterName, latestVersion),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_redis_cluster.by_version", 1),
				},
			},
			{
				// Force cluster deletion before final project destroy.
				Config: `
					resource "scaleway_account_project" "main" {}
				`,
				Check: waitForRedisClusterDeletion(tt, "scaleway_account_project.main", clusterName),
			},
		},
	})
}

func waitForRedisClusterDeletion(tt *acctest.TestTools, projectResourceName, clusterName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		projectResource, ok := state.RootModule().Resources[projectResourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", projectResourceName)
		}

		projectID := projectResource.Primary.ID
		redisAPI := redisSDK.NewAPI(meta.ExtractScwClient(tt.Meta))
		ctx := context.Background()

		return retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
			response, err := redisAPI.ListClusters(&redisSDK.ListClustersRequest{
				Zone:      "fr-par-1",
				ProjectID: &projectID,
				Name:      &clusterName,
			})
			if err != nil {
				return retry.NonRetryableError(err)
			}

			if len(response.Clusters) > 0 {
				return retry.RetryableError(fmt.Errorf("redis cluster %q still present in project %q", clusterName, projectID))
			}

			return nil
		})
	}
}
