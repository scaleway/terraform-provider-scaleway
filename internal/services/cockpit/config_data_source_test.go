package cockpit_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccCockpitConfig_DataSource_Basic(t *testing.T) {
	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_cockpit_config" "main" {
						region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_config.main", "region", "fr-par"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_metrics_retention.0.min_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_metrics_retention.0.max_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_metrics_retention.0.default_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_logs_retention.0.min_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_logs_retention.0.max_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_logs_retention.0.default_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_traces_retention.0.min_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_traces_retention.0.max_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "custom_traces_retention.0.default_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "product_metrics_retention.0.min_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "product_metrics_retention.0.max_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "product_metrics_retention.0.default_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "product_logs_retention.0.min_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "product_logs_retention.0.max_days"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_config.main", "product_logs_retention.0.default_days"),
				),
			},
		},
	})
}
