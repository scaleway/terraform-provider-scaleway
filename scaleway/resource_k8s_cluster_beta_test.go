package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
)

func TestAccScalewayK8SClusterMinimal(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterConfigMinimal("1.17.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "version", "1.17.3"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "tags.2", "minimal"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterConfigMinimal("1.17.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "version", "1.17.4"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.minimal", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "tags.2", "minimal"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterIngressDashboard(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterConfigIngressDashboard("1.18.0", "nginx", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "ingress", "nginx"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "enable_dashboard", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "tags.2", "ingressdashboard"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterConfigIngressDashboard("1.18.0", "traefik", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "ingress", "traefik"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "enable_dashboard", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.ingressdashboard", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.ingressdashboard", "tags.2", "ingressdashboard"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterAutoscaling(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterConfigAutoscaler("1.18.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.disable_scale_down", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.scale_down_delay_after_add", "20m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.estimator", "binpacking"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.expander", "most_pods"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.ignore_daemonsets_utilization", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.balance_similar_node_groups", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.expendable_pods_priority_cutoff", "0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.2", "autoscaler-config"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterAutoUpgrade(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(true, "tuesday", 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "tuesday"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "3"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(true, "any", 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(false, "any", 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists("scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
		},
	})
}

func testAccCheckScalewayK8SClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_k8s_cluster" {
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

func testAccCheckScalewayK8SClusterExists(n string) resource.TestCheckFunc {
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

func testAccCheckScalewayK8SClusterConfigMinimal(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "minimal" {
	cni = "calico"
	version = "%s"
	name = "minimal"
	tags = [ "terraform-test", "scaleway_k8s_cluster", "minimal" ]
}`, version)
}

func testAccCheckScalewayK8SClusterConfigIngressDashboard(version string, ingress string, dashboard bool) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "ingressdashboard" {
	cni = "calico"
	version = "%s"
	name = "ingress-dashboard"
	ingress = "%s"
	enable_dashboard = %t
	tags = [ "terraform-test", "scaleway_k8s_cluster", "ingressdashboard" ]
}`, version, ingress, dashboard)
}

func testAccCheckScalewayK8SClusterConfigAutoscaler(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "autoscaler" {
	cni = "calico"
	version = "%s"
	name = "autoscaler"
	autoscaler_config {
		disable_scale_down = true
		scale_down_delay_after_add = "20m"
		estimator = "binpacking"
		expander = "most_pods"
		ignore_daemonsets_utilization = true
		balance_similar_node_groups = true
		expendable_pods_priority_cutoff = 0
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster", "autoscaler-config" ]
}`, version)
}

func testAccCheckScalewayK8SClusterAutoUpgrade(enable bool, day string, hour uint64) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "auto_upgrade" {
	cni = "calico"
	version = "1.17.0"
	name = "default-pool"
	auto_upgrade {
	    enable = %t
		maintenance_window_start_hour = %d
		maintenance_window_day = "%s"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster", "auto_upgrade" ]
}`, enable, hour, day)
}
