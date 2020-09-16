package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceK8SCluster_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	dataSourceName := "data.scaleway_k8s_cluster.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_k8s_cluster_beta" "main" {
					  	name = "test%d"
					}
					
					data "scaleway_k8s_cluster" "main" {
					  	name = "${scaleway_k8s_cluster_beta.main.name}"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, dataSourceName),
				),
			},
		},
	})
}
