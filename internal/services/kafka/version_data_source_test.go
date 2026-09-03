package kafka_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceKafkaVersion_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	version := "4.0.0"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_kafka_version" "by_name" {
						name = %q
					}
				`, version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "name", version),
					resource.TestCheckResourceAttrSet("data.scaleway_kafka_version.by_name", "id"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "end_of_life_at", "2026-03-19T00:00:00Z"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.#", "2"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.0.name", "log_retention_bytes"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.0.hot_configurable", "true"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.0.int_property.min", "-1"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.0.int_property.default_value", "-1"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.0.int_property.unit", "bytes"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.1.name", "compression_type"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.by_name", "available_settings.1.string_property.default_value", "producer"),
				),
			},
		},
	})
}

func TestAccDataSourceKafkaVersion_Latest(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_kafka_version" "latest" {
						name = "latest"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_kafka_version.latest", "name", "4.0.0"),
					resource.TestCheckResourceAttrSet("data.scaleway_kafka_version.latest", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_kafka_version.latest", "end_of_life_at"),
				),
			},
		},
	})
}
