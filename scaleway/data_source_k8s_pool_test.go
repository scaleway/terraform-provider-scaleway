package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceK8SPool_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	dataSourceName := "data.scaleway_k8s_cluster.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_k8s_pool_beta" "main" {
					  	name = "test%d"
					}
					
					data "scaleway_k8s_pool" "main" {
					  	name = "${scaleway_k8s_pool.main.name}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SPoolExists(tt, dataSourceName),
				),
			},
		},
	})
}
