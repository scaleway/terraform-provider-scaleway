package redis_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceRedisClusterVersions_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_redis_cluster_versions" "redis" {
						include_disabled   = false
						include_beta       = false
						include_deprecated = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_redis_cluster_versions.redis", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_redis_cluster_versions.redis", "zone"),
					resource.TestCheckResourceAttrWith("data.scaleway_redis_cluster_versions.redis", "versions.#", func(value string) error {
						count, err := strconv.Atoi(value)
						if err != nil {
							return err
						}

						if count < 1 {
							return fmt.Errorf("expected at least one version, got %d", count)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("data.scaleway_redis_cluster_versions.redis", "versions.0.version"),
				),
			},
		},
	})
}
