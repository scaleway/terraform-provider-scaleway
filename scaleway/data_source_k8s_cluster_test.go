package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceK8SCluster_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	clusterName := "tf-cluster"
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
					  	tags    = [ "terraform-test", "data_scaleway_k8s_cluster", "basic" ]
					}`, clusterName, version),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_k8s_cluster" "main" {
					  	name 	= "%s"
						version = "%s"
						cni     = "cilium"
					  	tags    = [ "terraform-test", "data_scaleway_k8s_cluster", "basic" ]
					}
					
					data "scaleway_k8s_cluster" "prod" {
					  	name = "${scaleway_k8s_cluster.main.name}"
					}
					
					data "scaleway_k8s_cluster" "stg" {
					  	cluster_id = "${scaleway_k8s_cluster.main.id}"
					}`, clusterName, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "data.scaleway_k8s_cluster.prod"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_cluster.prod", "name", clusterName),
					testAccCheckScalewayK8SClusterExists(tt, "data.scaleway_k8s_cluster.stg"),
					resource.TestCheckResourceAttr("data.scaleway_k8s_cluster.stg", "name", clusterName),
				),
			},
		},
	})
}
