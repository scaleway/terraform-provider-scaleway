package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta4"
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

func TestAccScalewayK8SClusterBetaIngressDashboard(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigIngressDashboard("1.16.0", "nginx", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "version", "1.16.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "ingress", "nginx"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "enable_dashboard", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.max_size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "tags.2", "ingressdashboard"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigIngressDashboard("1.16.0", "traefik", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "version", "1.16.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "ingress", "traefik"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "enable_dashboard", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "default_pool.0.max_size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.ingressdashboard", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "tags.2", "ingressdashboard"),
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
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.expander", "most_pods"),
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
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.tags.2", "default-pool"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterBetaDefaultPoolWithPlacementGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigPoolWithPlacementGroup("1.16.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.pool_placement_group"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool_placement_group", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool_placement_group", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool_placement_group", "default_pool.0.placement_group_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool_placement_group", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool_placement_group", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool_placement_group", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool_placement_group", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool_placement_group", "tags.2", "default-pool-placement-group"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterBetaDefaultPoolRecreate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaDefaultPoolRecreate("gp1_xs"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.recreate_pool"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.recreate_pool", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.recreate_pool", "default_pool.0.node_type", "gp1_xs"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaDefaultPoolRecreate("gp1_s"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.recreate_pool", "default_pool.0.node_type", "gp1_s"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterBetaDefaultPoolWait(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigPoolWait("1.17.3", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.pool"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "version", "1.17.3"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "cni", "cilium"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.max_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.status", k8s.PoolStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.nodes.0.status", k8s.NodeStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.wait_for_pool_ready", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "tags.2", "default-pool"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigPoolWait("1.17.3", 2), // add a node and wait for the pool to be ready
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.pool"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.size", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.max_size", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.status", k8s.PoolStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.nodes.0.status", k8s.NodeStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.nodes.1.status", k8s.NodeStatusReady.String()), // check that the new node has the "ready" status
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.wait_for_pool_ready", "true"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigPoolWait("1.17.3", 1), // remove a node and wait for the pool to be ready
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.pool"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.pool", "default_pool.0.pool_id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.max_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.status", k8s.PoolStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.nodes.0.status", k8s.NodeStatusReady.String()),
					resource.TestCheckNoResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.nodes.1"), // check that the second node does not exist anymore
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.pool", "default_pool.0.wait_for_pool_ready", "true"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterBetaAutoUpgrade(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "tuesday", 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "tuesday"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "3"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "any", 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(false, "any", 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
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

		k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
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

		k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
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

func testAccCheckScalewayK8SClusterBetaConfigIngressDashboard(version string, ingress string, dashboard bool) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "ingressdashboard" {
	cni = "calico"
	version = "%s"
	name = "ingress-dashboard"
	ingress = "%s"
	enable_dashboard = %t
	default_pool {
		node_type = "gp1_xs"
		size = 1
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "ingressdashboard" ]
}`, version, ingress, dashboard)
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
		expander = "most_pods"
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
		tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "default-pool" ]
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "default-pool" ]
}`, version)
}

func testAccCheckScalewayK8SClusterBetaConfigPoolWait(version string, size int) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "pool" {
	cni = "cilium"
	version = "%s"
	name = "default-pool"
	default_pool {
		node_type = "gp1_xs"
		size = %d
		min_size = 1
		max_size = %d
		container_runtime = "docker"
		wait_for_pool_ready = true
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "default-pool" ]
}`, version, size, size)
}

func testAccCheckScalewayK8SClusterBetaConfigPoolWithPlacementGroup(version string) string {
	return fmt.Sprintf(`
resource "scaleway_instance_placement_group" "pool_placement_group" {
  name        = "pool-placement-group"
  policy_type = "max_availability"
  policy_mode = "optional"
}

resource "scaleway_k8s_cluster_beta" "pool_placement_group" {
	cni = "calico"
	version = "%s"
	name = "default-pool-placement-group"
	default_pool {
		node_type = "gp1_xs"
		size = 1
		placement_group_id = scaleway_instance_placement_group.pool_placement_group.id
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "default-pool-placement-group" ]
}`, version)
}

func testAccCheckScalewayK8SClusterBetaDefaultPoolRecreate(nodeType string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "recreate_pool" {
	cni = "calico"
	version = "1.17.0"
	name = "default-pool"
	default_pool {
		node_type = "%s"
		size = 1
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "recreate-pool" ]
}`, nodeType)
}

func testAccCheckScalewayK8SClusterBetaAutoUpgrade(enable bool, day string, hour uint64) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "auto_upgrade" {
	cni = "calico"
	version = "1.17.0"
	name = "default-pool"
	default_pool {
		node_type = "gp1_xs"
		size = 1
	}
	auto_upgrade {
	    enable = %t
		maintenance_window_start_hour = %d
		maintenance_window_day = "%s"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "auto_upgrade" ]
}`, enable, hour, day)
}
