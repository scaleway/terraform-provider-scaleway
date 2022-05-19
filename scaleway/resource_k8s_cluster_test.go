package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_k8s_cluster", &resource.Sweeper{
		Name: "scaleway_k8s_cluster",
		F:    testSweepK8SCluster,
	})
}

func testAccScalewayK8SClusterGetLatestK8SVersion(tt *TestTools) string {
	api := k8s.NewAPI(tt.Meta.scwClient)
	versions, err := api.ListVersions(&k8s.ListVersionsRequest{})
	if err != nil {
		tt.T.Fatalf("Could not get latestK8SVersion: %s", err)
	}
	if len(versions.Versions) > 1 {
		latestK8SVersion := versions.Versions[0].Name
		return latestK8SVersion
	}
	return ""
}
func testAccScalewayK8SClusterGetLatestK8SVersionMinor(tt *TestTools) string {
	api := k8s.NewAPI(tt.Meta.scwClient)
	versions, err := api.ListVersions(&k8s.ListVersionsRequest{})
	if err != nil {
		tt.T.Fatalf("Could not get latestK8SVersion: %s", err)
	}
	if len(versions.Versions) > 1 {
		latestK8SVersion := versions.Versions[0].Name
		latestK8SVersionMinor, _ := k8sGetMinorVersionFromFull(latestK8SVersion)
		return latestK8SVersionMinor
	}
	return ""
}

func testAccScalewayK8SClusterGetPreviousK8SVersion(tt *TestTools) string {
	api := k8s.NewAPI(tt.Meta.scwClient)
	versions, err := api.ListVersions(&k8s.ListVersionsRequest{})
	if err != nil {
		tt.T.Fatalf("Could not get latestK8SVersion: %s", err)
	}
	if len(versions.Versions) > 1 {
		previousK8SVersion := versions.Versions[1].Name
		return previousK8SVersion
	}
	return ""
}

func testAccScalewayK8SClusterGetPreviousK8SVersionMinor(tt *TestTools) string {
	api := k8s.NewAPI(tt.Meta.scwClient)
	versions, err := api.ListVersions(&k8s.ListVersionsRequest{})
	if err != nil {
		tt.T.Fatalf("Could not get latestK8SVersion: %s", err)
	}
	if len(versions.Versions) > 1 {
		previousK8SVersion := versions.Versions[1].Name
		previousK8SVersionMinor, _ := k8sGetMinorVersionFromFull(previousK8SVersion)
		return previousK8SVersionMinor
	}
	return ""
}

func testSweepK8SCluster(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms}, func(scwClient *scw.Client, region scw.Region) error {
		k8sAPI := k8s.NewAPI(scwClient)

		l.Debugf("sweeper: destroying the k8s cluster in (%s)", region)
		listClusters, err := k8sAPI.ListClusters(&k8s.ListClustersRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing clusters in (%s) in sweeper: %w", region, err)
		}

		for _, cluster := range listClusters.Clusters {
			//remove pools
			listPools, err := k8sAPI.ListPools(&k8s.ListPoolsRequest{
				Region:    region,
				ClusterID: cluster.ID,
			}, scw.WithAllPages())
			if err != nil {
				return fmt.Errorf("error listing pool in (%s) in sweeper: %w", region, err)
			}

			for _, pool := range listPools.Pools {
				_, err := k8sAPI.DeletePool(&k8s.DeletePoolRequest{
					Region: region,
					PoolID: pool.ID,
				})
				if err != nil {
					return fmt.Errorf("error deleting pool in sweeper: %w", err)
				}
			}
			_, err = k8sAPI.DeleteCluster(&k8s.DeleteClusterRequest{
				Region:    region,
				ClusterID: cluster.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting cluster in sweeper: %w", err)
			}
		}

		return nil
	})
}

func TestAccScalewayK8SCluster_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	latestK8SVersion := testAccScalewayK8SClusterGetLatestK8SVersion(tt)
	previousK8SVersion := testAccScalewayK8SClusterGetPreviousK8SVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterConfigMinimal(previousK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "version", previousK8SVersion),
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
				Config: testAccCheckScalewayK8SClusterConfigMinimal(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.minimal", "version", latestK8SVersion),
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

func TestAccScalewayK8SCluster_Autoscaling(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	latestK8SVersion := testAccScalewayK8SClusterGetLatestK8SVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterConfigAutoscaler(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "version", latestK8SVersion),
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
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.scale_down_unneeded_time", "20m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.estimator", "binpacking"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.expander", "most_pods"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.ignore_daemonsets_utilization", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.balance_similar_node_groups", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.expendable_pods_priority_cutoff", "10"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.scale_down_utilization_threshold", "0.77"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.max_graceful_termination_sec", "1337"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.2", "autoscaler-config"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterConfigAutoscalerChange(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.autoscaler"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "version", latestK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "cni", "calico"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.autoscaler", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.disable_scale_down", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.scale_down_delay_after_add", "20m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.scale_down_unneeded_time", "5m"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.estimator", "binpacking"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.expander", "most_pods"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.ignore_daemonsets_utilization", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.balance_similar_node_groups", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.expendable_pods_priority_cutoff", "0"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.scale_down_utilization_threshold", "0.33"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "autoscaler_config.0.max_graceful_termination_sec", "2664"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.autoscaler", "tags.2", "autoscaler-config"),
				),
			},
		},
	})
}

func TestAccScalewayK8SCluster_OIDC(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	latestK8SVersion := testAccScalewayK8SClusterGetLatestK8SVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterConfigOIDC(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.oidc"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "version", latestK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "cni", "cilium"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "status", k8s.ClusterStatusPoolRequired.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.issuer_url", "https://api.scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.client_id", "my-super-id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.username_claim", "mario"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.groups_prefix", "pouf"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.groups_claim.0", "k8s"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.groups_claim.1", "admin"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "tags.2", "oidc-config"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterConfigOIDCChange(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.oidc"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "version", latestK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "cni", "cilium"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "status", k8s.ClusterStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.config_file"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.cluster_ca_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "kubeconfig.0.token"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "apiserver_url"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_cluster.oidc", "wildcard_dns"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.issuer_url", "https://secretapi.scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.client_id", "my-even-more-awesome-id"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.username_claim", "luigi"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.username_prefix", "boo"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "open_id_connect_config.0.groups_prefix", ""),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.oidc", "tags.2", "oidc-config"),
				),
			},
		},
	})
}

func TestAccScalewayK8SCluster_AutoUpgrade(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	latestK8SVersion := testAccScalewayK8SClusterGetLatestK8SVersion(tt)
	latestK8SVersionMinor := testAccScalewayK8SClusterGetLatestK8SVersionMinor(tt)
	previousK8SVersion := testAccScalewayK8SClusterGetPreviousK8SVersion(tt)
	previousK8SVersionMinor := testAccScalewayK8SClusterGetPreviousK8SVersionMinor(tt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(false, "any", 0, previousK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "version", previousK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(true, "any", 0, previousK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "version", previousK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(true, "any", 0, latestK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "version", latestK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(false, "any", 0, latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "version", latestK8SVersion),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(true, "tuesday", 3, latestK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "version", latestK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "tuesday"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "3"),
				),
			},
			{
				Config: testAccCheckScalewayK8SClusterAutoUpgrade(true, "any", 0, latestK8SVersionMinor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.auto_upgrade"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "version", latestK8SVersionMinor),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.enable", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_day", "any"),
					resource.TestCheckResourceAttr("scaleway_k8s_cluster.auto_upgrade", "auto_upgrade.0.maintenance_window_start_hour", "0"),
				),
			},
		},
	})
}

func testAccCheckScalewayK8SClusterDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_k8s_cluster" {
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

func testAccCheckScalewayK8SClusterExists(tt *TestTools, n string) resource.TestCheckFunc {
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

func testAccCheckScalewayK8SClusterConfigMinimal(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "minimal" {
	cni = "calico"
	version = "%s"
	name = "ClusterConfigMinimal"
	tags = [ "terraform-test", "scaleway_k8s_cluster", "minimal" ]
}`, version)
}

func testAccCheckScalewayK8SClusterConfigAutoscaler(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "autoscaler" {
	cni = "calico"
	version = "%s"
	name = "autoscaler-01"
	region = "nl-ams"
	autoscaler_config {
		disable_scale_down = true
		scale_down_delay_after_add = "20m"
		scale_down_unneeded_time = "20m"
		estimator = "binpacking"
		expander = "most_pods"
		ignore_daemonsets_utilization = true
		balance_similar_node_groups = true
		expendable_pods_priority_cutoff = 10
		scale_down_utilization_threshold = 0.77
		max_graceful_termination_sec = 1337
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster", "autoscaler-config" ]
}`, version)
}

func testAccCheckScalewayK8SClusterConfigAutoscalerChange(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "autoscaler" {
	cni = "calico"
	version = "%s"
	name = "autoscaler-02"
	region = "nl-ams"
	autoscaler_config {
		disable_scale_down = false
		scale_down_delay_after_add = "20m"
		scale_down_unneeded_time = "5m"
		estimator = "binpacking"
		expander = "most_pods"
		expendable_pods_priority_cutoff = 0
		scale_down_utilization_threshold = 0.33
		max_graceful_termination_sec = 2664
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster", "autoscaler-config" ]
}`, version)
}

func testAccCheckScalewayK8SClusterConfigOIDC(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "oidc" {
	cni = "cilium"
	version = "%s"
	name = "oidc"
	open_id_connect_config {
		issuer_url = "https://api.scaleway.com"
		client_id = "my-super-id"
		username_claim = "mario"
		groups_claim = [ "k8s", "admin" ]
		groups_prefix = "pouf"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster", "oidc-config" ]
}

resource "scaleway_k8s_pool" "minimal" {
    name = "minimal"
	cluster_id = "${scaleway_k8s_cluster.oidc.id}"
	node_type = "gp1_xs"
	autohealing = true
	autoscaling = true
	size = 1
	tags = [ "terraform-test", "scaleway_k8s_cluster", "minimal" ]
}
`, version)
}

func testAccCheckScalewayK8SClusterConfigOIDCChange(version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "oidc" {
	cni = "cilium"
	version = "%s"
	name = "oidc"
	open_id_connect_config {
		issuer_url = "https://secretapi.scaleway.com"
		client_id = "my-even-more-awesome-id"
		username_claim = "luigi"
		groups_claim = [ ]
		username_prefix = "boo"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster", "oidc-config" ]
}

resource "scaleway_k8s_pool" "oidc" {
    name = "minimal"
	cluster_id = "${scaleway_k8s_cluster.oidc.id}"
	node_type = "gp1_xs"
	autohealing = true
	autoscaling = true
	size = 1
	tags = [ "terraform-test", "scaleway_k8s_cluster", "minimal" ]
}
`, version)
}

func testAccCheckScalewayK8SClusterAutoUpgrade(enable bool, day string, hour uint64, version string) string {
	return fmt.Sprintf(`
resource "scaleway_k8s_cluster" "auto_upgrade" {
	cni = "calico"
	version = "%s"
	name = "default-pool"
	auto_upgrade {
	    enable = %t
		maintenance_window_start_hour = %d
		maintenance_window_day = "%s"
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster", "auto_upgrade" ]
}`, version, enable, hour, day)
}
