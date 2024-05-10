package scw_config_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceConfig_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_config" "main" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_config.main", "project_id", "11111111-2222-3333-4444-555555555555"),
					resource.TestCheckResourceAttr("data.scaleway_config.main", "access_key", "SCWXXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr("data.scaleway_config.main", "secret_key", "01234567-abcd-effe-dcba-012345678910"),
				),
			},
		},
	})
}
