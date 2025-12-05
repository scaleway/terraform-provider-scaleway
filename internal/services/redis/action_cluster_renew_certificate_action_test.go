package redis_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionRedisClusterRenewCertificate_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRedisClusterRenewCertificate_Basic because actions are not yet supported on OpenTofu")
	}

	// Note: This test will fail with 501 Not Implemented until Scaleway implements
	// the RenewClusterCertificate API endpoint. The SDK method exists but the API
	// endpoint is not yet available on the server side.
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestRedisVersion := getLatestVersion(tt)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_redis_cluster" "main" {
						name         = "test-redis-action-renew-certificate"
						version      = "%s"
						node_type    = "RED1-XS"
						user_name    = "my_initial_user"
						password     = "thiZ_is_v&ry_s3cret"
						cluster_size = 1
						tls_enabled  = "true"
						zone         = "fr-par-2"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_redis_cluster_renew_certificate_action.main]
							}
						}
					}

					action "scaleway_redis_cluster_renew_certificate_action" "main" {
						config {
							cluster_id = scaleway_redis_cluster.main.id
							wait       = true
						}
					}
				`, latestRedisVersion),
			},
		},
	})
}
