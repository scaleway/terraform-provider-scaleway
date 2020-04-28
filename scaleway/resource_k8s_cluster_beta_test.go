package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_k8s_cluster_beta", &resource.Sweeper{
		Name: "scaleway_k8s_cluster_beta",
		F:    testSweepK8SCluster,
	})
}

func testSweepK8SCluster(region string) error {
	scwClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	k8sAPI := k8s.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the k8s cluster in (%s)", region)
	listClusters, err := k8sAPI.ListClusters(&k8s.ListClustersRequest{}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing clusters in (%s) in sweeper: %s", region, err)
	}

	for _, cluster := range listClusters.Clusters {
		_, err := k8sAPI.DeleteCluster(&k8s.DeleteClusterRequest{
			ClusterID: cluster.ID,
		})
		if err != nil {
			return fmt.Errorf("error deleting cluster in sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayK8SClusterDeprecated(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaDeprecated("1.17.4", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.deprecated"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "version", "1.17.4"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "tags.2", "deprecated"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "default_pool.0.size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "default_pool.0.node_type", "dev1_m"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaDeprecated("1.17.4", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.deprecated"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "version", "1.17.4"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "tags.2", "deprecated"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "default_pool.0.size", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "default_pool.0.node_type", "dev1_m"),
				),
			},
		},
	})

}

func TestAccScalewayK8SClusterMinimal(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigMinimal("1.17.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "version", "1.17.3"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "status", k8s.ClusterStatusPoolRequired.String()),
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
				Config: testAccCheckScalewayK8SClusterBetaConfigMinimal("1.17.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "version", "1.17.4"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "status", k8s.ClusterStatusPoolRequired.String()),
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

func TestAccScalewayK8SClusterIngressDashboard(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigIngressDashboard("1.18.0", "nginx", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "ingress", "nginx"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "enable_dashboard", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "status", k8s.ClusterStatusPoolRequired.String()),
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
				Config: testAccCheckScalewayK8SClusterBetaConfigIngressDashboard("1.18.0", "traefik", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "ingress", "traefik"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "enable_dashboard", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "status", k8s.ClusterStatusPoolRequired.String()),
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

func TestAccScalewayK8SClusterAutoscaling(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigAutoscaler("1.18.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.disable_scale_down", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.scale_down_delay_after_add", "20m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.scale_down_unneeded_time", "20m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.estimator", "binpacking"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.expander", "most_pods"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.ignore_daemonsets_utilization", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.balance_similar_node_groups", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.expendable_pods_priority_cutoff", "10"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.2", "autoscaler-config"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigAutoscalerChange("1.18.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster_beta.autoscaler", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.disable_scale_down", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.scale_down_delay_after_add", "20m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.scale_down_unneeded_time", "5m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.estimator", "binpacking"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.expander", "most_pods"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.ignore_daemonsets_utilization", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.balance_similar_node_groups", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "autoscaler_config.0.expendable_pods_priority_cutoff", "0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.1", "scaleway_k8s_cluster_beta"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "tags.2", "autoscaler-config"),
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterAutoUpgrade(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(false, "any", 0, "1.18.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", "1.18.0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "tuesday", 3, "1.18"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", "1.18"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "tuesday"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "3"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "any", 0, "1.18"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", "1.18"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
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

func testAccCheckScalewayK8SClusterBetaDeprecated(version string, size int) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "deprecated" {
	cni = "calico"
	version = "%s"
	name = "deprecated"
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "deprecated" ]
	default_pool {
	  node_type = "DEV1-M"
	  size = %d
	}
}`, version, size)
}

func testAccCheckScalewayK8SClusterBetaConfigMinimal(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "minimal" {
	cni = "calico"
	version = "%s"
	name = "minimal"
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
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "ingressdashboard" ]
}`, version, ingress, dashboard)
}

func testAccCheckScalewayK8SClusterBetaConfigAutoscaler(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "autoscaler" {
	cni = "calico"
	version = "%s"
	name = "autoscaler"
	autoscaler_config {
		disable_scale_down = true
		scale_down_delay_after_add = "20m"
		scale_down_unneeded_time = "20m"
		estimator = "binpacking"
		expander = "most_pods"
		ignore_daemonsets_utilization = true
		balance_similar_node_groups = true
		expendable_pods_priority_cutoff = 10
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "autoscaler-config" ]
}`, version)
}

func testAccCheckScalewayK8SClusterBetaConfigAutoscalerChange(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "autoscaler" {
	cni = "calico"
	version = "%s"
	name = "autoscaler"
	autoscaler_config {
		disable_scale_down = false
		scale_down_delay_after_add = "20m"
		scale_down_unneeded_time = "5m"
		estimator = "binpacking"
		expander = "most_pods"
		expendable_pods_priority_cutoff = 0
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "autoscaler-config" ]
}`, version)
}

func testAccCheckScalewayK8SClusterBetaAutoUpgrade(enable bool, day string, hour uint64, version string) string {
	if enable {
		return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "auto_upgrade" {
	cni = "calico"
	version = "%s"
	name = "default-pool"
	auto_upgrade {
	    enable = true
		maintenance_window_start_hour = %d
		maintenance_window_day = "%s"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "auto_upgrade" ]
}`, version, hour, day)
	}
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "auto_upgrade" {
	cni = "calico"
	version = "%s"
	name = "default-pool"
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "auto_upgrade" ]
}`, version)
}
