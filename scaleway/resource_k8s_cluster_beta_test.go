package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var (
	latestK8SVersion        = "1.19.2"
	latestK8SVersionMinor   = "1.19"
	previousK8SVersion      = "1.18.9"
	previousK8SVersionMinor = "1.18"
)

func init() {
	resource.AddTestSweepers("scaleway_k8s_cluster_beta", &resource.Sweeper{
		Name: "scaleway_k8s_cluster_beta",
		F:    testSweepK8SCluster,
	})
}

func testAccScalewayK8SClusterGetLatestVersion(tt *TestTools) {
	api := k8s.NewAPI(tt.Meta.scwClient)
	versions, err := api.ListVersions(&k8s.ListVersionsRequest{})
	if err != nil {
		tt.T.Fatalf("Could not get latestK8SVersion: %s", err)
		return
	}
	if len(versions.Versions) > 1 {
		latestK8SVersion = versions.Versions[0].Name
		latestK8SVersionMinor, _ = k8sGetMinorVersionFromFull(latestK8SVersion)
		previousK8SVersion = versions.Versions[1].Name
		previousK8SVersionMinor, _ = k8sGetMinorVersionFromFull(previousK8SVersion)
	}
}

func testSweepK8SCluster(region string) error {
	scwClient, err := sharedClientForRegion(scw.Region(region))
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

func TestAccScalewayK8SCluster_Deprecated(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccScalewayK8SClusterGetLatestVersion(tt)
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaDeprecated(latestK8SVersion, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.deprecated"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "version", latestK8SVersion),
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
				Config: testAccCheckScalewayK8SClusterBetaDeprecated(latestK8SVersion, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.deprecated"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.deprecated", "version", latestK8SVersion),
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

func TestAccScalewayK8SCluster_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccScalewayK8SClusterGetLatestVersion(tt)
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigMinimal(previousK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "version", previousK8SVersion),
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
				Config: testAccCheckScalewayK8SClusterBetaConfigMinimal(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.minimal", "version", latestK8SVersion),
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

func TestAccScalewayK8SCluster_IngressDashboard(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccScalewayK8SClusterGetLatestVersion(tt)
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigIngressDashboard(latestK8SVersion, "nginx", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "version", latestK8SVersion),
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
				Config: testAccCheckScalewayK8SClusterBetaConfigIngressDashboard(latestK8SVersion, "traefik", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "version", latestK8SVersion),
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
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigIngressDashboard(latestK8SVersion, "traefik2", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.ingressdashboard"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "version", latestK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.ingressdashboard", "ingress", "traefik2"),
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

func TestAccScalewayK8SCluster_Autoscaling(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccScalewayK8SClusterGetLatestVersion(tt)
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaConfigAutoscaler(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "version", latestK8SVersion),
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
				Config: testAccCheckScalewayK8SClusterBetaConfigAutoscalerChange(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.autoscaler", "version", latestK8SVersion),
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

func TestAccScalewayK8SCluster_AutoUpgrade(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccScalewayK8SClusterGetLatestVersion(tt)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(false, "any", 0, previousK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", previousK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "any", 0, previousK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", previousK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "any", 0, latestK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", latestK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(false, "any", 0, latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", latestK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "tuesday", 3, latestK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", latestK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "tuesday"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "3"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterBetaAutoUpgrade(true, "any", 0, latestK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists(tt, "scaleway_k8s_cluster_beta.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "version", latestK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster_beta.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
		},
	})
}

func testAccCheckScalewayK8SClusterBetaDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_k8s_cluster_beta" {
				continue
			}

			k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
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
}

func testAccCheckScalewayK8SClusterBetaExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
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
	name = "ClusterBetaDeprecated"
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
	name = "ClusterBetaConfigMinimal"
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
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "auto_upgrade" {
	cni = "calico"
	version = "%s"
	name = "default-pool"
	auto_upgrade {
	    enable = %t
		maintenance_window_start_hour = %d
		maintenance_window_day = "%s"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "auto_upgrade" ]
}`, version, enable, hour, day)
}
