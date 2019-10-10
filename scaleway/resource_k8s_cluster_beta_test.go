package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta3"
)

func TestAccScalewayK8SClusterBetaMinimal(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigMinimal("1.16.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "version", "1.16.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.max_size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "tags.2", "minimal"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigMinimal("1.16.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "version", "1.16.1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "default_pool.0.max_size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.minimal", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "tags.2", "minimal"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterBetaAutoscaling(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigAutoscaler("1.16.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "version", "1.16.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "default_pool.0.max_size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.disable_scale_down", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.scale_down_delay_after_add", "20m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.estimator", "binpacking"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.expander", "most-pods"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.ignore_daemonsets_utilization", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.balance_similar_node_groups", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.expendable_pods_priority_cutoff", "0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.2", "autoscaler-config"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterBetaDefaultPool(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigPool("1.16.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.pool"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "version", "1.16.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.max_size", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.autoscaling", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.autohealing", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.container_runtime", "docker"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "tags.2", "default-pool"),
				),
			},
		},
	})
}

func testAccCheckScalewayK8SClusterBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_k8s_cluster_beta" {
			continue
		}

		k8sAPI, region, clusterID, err := getK8SAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = k8sAPI.GetCluster(&k8s.GetClusterRequest{
			Region:    region,
			ClusterID: clusterID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("cluster (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}
	return nil
}

func testAccCheckScalewayK8SClusterBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		k8sAPI, region, clusterID, err := getK8SAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = k8sAPI.GetCluster(&k8s.GetClusterRequest{
			Region:    region,
			ClusterID: clusterID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayK8SClusterBetaConfigMinimal(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "minimal" {
	cni = "calico"
	version = "%s"
	name = "minimal"
	default_pool {
		node_type = "gp1_xs"
		size = 1
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "minimal" ]
}`, version)
}

func testAccCheckScalewayK8SClusterBetaConfigAutoscaler(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "autoscaler" {
	cni = "calico"
	version = "%s"
	name = "autoscaler"
	default_pool {
		node_type = "gp1_xs"
		size = 1
	}
	autoscaler_config {
		disable_scale_down = true
		scale_down_delay_after_add = "20m"
		estimator = "binpacking"
		expander = "most-pods"
		ignore_daemonsets_utilization = true
		balance_similar_node_groups = true
		expendable_pods_priority_cutoff = 0
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "autoscaler-config" ]
}`, version)
}

func testAccCheckScalewayK8SClusterBetaConfigPool(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "pool" {
	cni = "calico"
	version = "%s"
	name = "default-pool"
	default_pool {
		node_type = "gp1_xs"
		size = 1
		min_size = 1
		max_size = 2
		autoscaling = true
		autohealing = true
		container_runtime = "docker"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "default-pool" ]
}`, version)
}
