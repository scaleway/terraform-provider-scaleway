package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceK8SPool_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	clusterName := "tf-cluster-pool"
	poolName := "tf-pool"
	version := testAccScalewayK8SClusterGetLatestK8SVersion(tt)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_k8s_cluster" "main" {
					  	name 	= "%s"
						version = "%s"
						cni     = "cilium"
					  	tags    = [ "terraform-test", "data_scaleway_k8s_pool", "basic" ]
					}
					
					resource "scaleway_k8s_pool" "default" {
						cluster_id = "${scaleway_k8s_cluster.main.id}"
						name = "%s"
						node_type = "dev1_m"
						size = 1
						tags = [ "terraform-test", "data_scaleway_k8s_pool", "basic" ]
					}`, clusterName, version, poolName),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_k8s_cluster" "main" {
					  	name 	= "%s"
						version = "%s"
						cni     = "cilium"
					  	tags    = [ "terraform-test", "data_scaleway_k8s_cluster", "basic" ]
					}

					resource "scaleway_k8s_pool" "default" {
						cluster_id = "${scaleway_k8s_cluster.main.id}"
						name = "%s"
						node_type = "dev1_m"
						size = 1
						tags = [ "terraform-test", "data_scaleway_k8s_pool", "basic" ]
					}
					
					data "scaleway_k8s_pool" "prod" {
					  	name = "${scaleway_k8s_pool.default.name}"
						cluster_id = "${scaleway_k8s_cluster.main.id}"
					}
					
					data "scaleway_k8s_pool" "stg" {
					  	pool_id = "${scaleway_k8s_pool.default.id}"
					}`, clusterName, version, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SPoolExists(tt, "data.scaleway_k8s_pool.prod"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_pool.prod", "name", poolName),
					testAccCheckScalewayK8SPoolExists(tt, "data.scaleway_k8s_pool.stg"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_pool.stg", "name", poolName),
				),
			},
		},
	})
}
