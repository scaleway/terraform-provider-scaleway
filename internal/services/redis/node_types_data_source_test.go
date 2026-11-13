package redis_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceRedisNodeTypes_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_redis_node_types" "available" {
						include_disabled_types = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_redis_node_types.available", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_redis_node_types.available", "zone"),
					resource.TestCheckResourceAttrWith("data.scaleway_redis_node_types.available", "node_types.#", func(value string) error {
						count, err := strconv.Atoi(value)
						if err != nil {
							return err
						}

						if count < 1 {
							return fmt.Errorf("expected at least one node type, got %d", count)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("data.scaleway_redis_node_types.available", "node_types.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_redis_node_types.available", "node_types.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.scaleway_redis_node_types.available", "node_types.0.memory_size_in_gb"),
					resource.TestCheckResourceAttrSet("data.scaleway_redis_node_types.available", "node_types.0.stock_status"),
				),
			},
		},
	})
}
