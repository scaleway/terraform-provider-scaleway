package redis_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	redisSDK "github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func TestAccActionRedisClusterRenewCertificate_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRedisClusterRenewCertificate_Basic because action are not yet supported on OpenTofu")
	}

	// API documents RenewClusterCertificate, but the backend currently returns HTTP 501 Not Implemented
	// (reproduced 2026-10-09 on fr-par-1). Re-enable and record a cassette once the endpoint works.
	t.Skip("Skipping until Redis RenewClusterCertificate API is implemented (currently returns 501 Not Implemented)")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestRedisVersion := getLatestVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isClusterDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_redis_cluster" "main" {
						name = "tf-tests-redis-action-renew-cert"
						version = %q
						node_type = "RED1-XS"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						cluster_size = 1
						tls_enabled = true

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_redis_cluster_renew_certificate.main]
							}
						}
					}

					action "scaleway_redis_cluster_renew_certificate" "main" {
						config {
							cluster_id = scaleway_redis_cluster.main.id
							wait = true
						}
					}
				`, latestRedisVersion),
				Check: resource.ComposeTestCheckFunc(
					isClusterCertificateRenewed(tt, "scaleway_redis_cluster.main"),
				),
			},
		},
	})
}

func isClusterCertificateRenewed(tt *acctest.TestTools, clusterResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[clusterResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", clusterResourceName)
		}

		zone, id, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse cluster ID: %w", err)
		}

		api := redisSDK.NewAPI(tt.Meta.ScwClient())

		var cluster *redisSDK.Cluster

		err = transport.RetryOn403(tt.T.Context(), func() error {
			var getErr error

			cluster, getErr = api.GetCluster(&redisSDK.GetClusterRequest{
				Zone:      zone,
				ClusterID: id,
			}, scw.WithContext(tt.T.Context()))

			return getErr
		})
		if err != nil {
			return fmt.Errorf("failed to get cluster: %w", err)
		}

		if cluster == nil {
			return fmt.Errorf("cluster %s not found", id)
		}

		if cluster.Status != redisSDK.ClusterStatusReady {
			return fmt.Errorf("cluster %s is not ready after certificate renewal, status: %s", id, cluster.Status)
		}

		return nil
	}
}
