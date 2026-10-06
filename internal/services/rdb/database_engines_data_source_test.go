package rdb_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceRDBDatabaseEngines_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_rdb_database_engines" "pg" {
						name = "PostgreSQL"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database_engines.pg", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database_engines.pg", "region"),
					resource.TestCheckResourceAttrWith("data.scaleway_rdb_database_engines.pg", "engines.#", func(value string) error {
						count, err := strconv.Atoi(value)
						if err != nil {
							return err
						}

						if count < 1 {
							return fmt.Errorf("expected at least one engine, got %d", count)
						}

						return nil
					}),
					resource.TestCheckResourceAttr("data.scaleway_rdb_database_engines.pg", "engines.0.name", "PostgreSQL"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database_engines.pg", "engines.0.versions.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database_engines.pg", "engines.0.versions.0.version"),
				),
			},
		},
	})
}
