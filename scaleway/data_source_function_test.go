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
						name = "test-cr-data"
					}
					
					data "scaleway_function" "by_name" {
						name = scaleway_function.main.name
					}
					
					data "scaleway_function" "by_id" {
						function_id = scaleway_function.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "test-cr-data"),
					resource.TestCheckResourceAttrSet("data.scaleway_function_namespace.by_name", "id"),

					resource.TestCheckResourceAttr("data.scaleway_function_namespace.by_id", "name", "test-cr-data"),
					resource.TestCheckResourceAttrSet("data.scaleway_function_namespace.by_id", "id"),
				),
			},
		},
	})
}
