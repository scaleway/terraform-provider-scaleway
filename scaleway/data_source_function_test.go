package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceFunction_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_function_namespace" "main" {
						name = "tf-ds-function"
					}

					resource scaleway_function main {
						name = "tf-ds-function"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node14"
						privacy = "private"
						handler = "handler.handle"
					}					
					
					data "scaleway_function" "by_name" {
						name = scaleway_function.main.name
						namespace_id = scaleway_function_namespace.main.id
					}
					
					data "scaleway_function" "by_id" {
						function_id = scaleway_function.main.id
						namespace_id = scaleway_function_namespace.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionExists(tt, "scaleway_function.main"),
					resource.TestCheckResourceAttr("scaleway_function.main", "name", "tf-ds-function"),
					resource.TestCheckResourceAttrSet("data.scaleway_function.by_name", "id"),

					resource.TestCheckResourceAttr("data.scaleway_function.by_id", "name", "tf-ds-function"),
					resource.TestCheckResourceAttrSet("data.scaleway_function.by_id", "id"),
					resource.TestCheckResourceAttr("data.scaleway_function.by_id", "name", "tf-ds-function"),
				),
			},
		},
	})
}
