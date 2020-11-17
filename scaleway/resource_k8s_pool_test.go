package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
)

func TestAccScalewayK8SCluster_PoolBasic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	latestK8SVersion := testAccScalewayK8SClusterGetLatestK8SVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SPoolConfigMinimal(latestK8SVersion, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolExists(tt, "scaleway_k8s_pool.default"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "autohealing", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "autoscaling", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "version", latestK8SVersion),
					resource.TestCheckResourceAttrSet("scaleway_k8s_pool.default", "id"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "tags.2", "default"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolConfigMinimal(latestK8SVersion, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolExists(tt, "scaleway_k8s_pool.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "autohealing", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "autoscaling", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "version", latestK8SVersion),
					resource.TestCheckResourceAttrSet("scaleway_k8s_pool.minimal", "id"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "tags.1", "scaleway_k8s_cluster"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "tags.2", "minimal"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolConfigMinimal(latestK8SVersion, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolBetaDestroy(tt, "scaleway_k8s_pool.minimal"),
				),
			},
		},
	})
}

func TestAccScalewayK8SCluster_PoolWait(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	latestK8SVersion := testAccScalewayK8SClusterGetLatestK8SVersion(tt)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SPoolConfigWait(latestK8SVersion, false, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolExists(tt, "scaleway_k8s_pool.default"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "max_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "status", k8s.PoolStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "nodes.0.status", k8s.NodeStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.default", "wait_for_pool_ready", "true"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolConfigWait(latestK8SVersion, true, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolExists(tt, "scaleway_k8s_pool.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "max_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "status", k8s.PoolStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "nodes.0.status", k8s.NodeStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "wait_for_pool_ready", "true"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolConfigWait(latestK8SVersion, true, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolExists(tt, "scaleway_k8s_pool.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "size", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "max_size", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "status", k8s.PoolStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "nodes.0.status", k8s.NodeStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "nodes.1.status", k8s.NodeStatusReady.String()), // check that the new node has the "ready" status
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "wait_for_pool_ready", "true"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolConfigWait(latestK8SVersion, true, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolExists(tt, "scaleway_k8s_pool.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "min_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "max_size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "status", k8s.PoolStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "nodes.0.status", k8s.NodeStatusReady.String()),
					resource.TestCheckNoResourceAttr("scaleway_k8s_pool.minimal", "nodes.1"), // check that the second node does not exist anymore
					resource.TestCheckResourceAttr("scaleway_k8s_pool.minimal", "wait_for_pool_ready", "true"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolConfigWait(latestK8SVersion, false, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.minimal"),
					testAccCheckScalewayK8SPoolBetaDestroy(tt, "scaleway_k8s_pool.minimal"),
				),
			},
		},
	})
}
func TestAccScalewayK8SCluster_PoolPlacementGroup(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	latestK8SVersion := testAccScalewayK8SClusterGetLatestK8SVersion(tt)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SPoolConfigPlacementGroup(latestK8SVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterExists(tt, "scaleway_k8s_cluster.placement_group"),
					testAccCheckScalewayK8SPoolExists(tt, "scaleway_k8s_pool.placement_group"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.placement_group", "node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool.placement_group", "size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_pool.placement_group", "id"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_pool.placement_group", "placement_group_id"),
				),
			},
		},
	})
}

func testAccCheckScalewayK8SPoolBetaDestroy(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return nil
		}

		k8sAPI, region, poolID, err := k8sAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = k8sAPI.GetPool(&k8s.GetPoolRequest{
			Region: region,
			PoolID: poolID,
		})
		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("pool (%s) still exists", rs.Primary.ID)
		}
		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayK8SPoolExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		k8sAPI, region, poolID, err := k8sAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = k8sAPI.GetPool(&k8s.GetPoolRequest{
			Region: region,
			PoolID: poolID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayK8SPoolConfigMinimal(version string, otherPool bool) string {
	pool := ""
	if otherPool {
		pool += `
resource "scaleway_k8s_pool" "minimal" {
    name = "minimal"
	cluster_id = "${scaleway_k8s_cluster.minimal.id}"
	node_type = "gp1_xs"
	autohealing = true
	autoscaling = true
	size = 1
	tags = [ "terraform-test", "scaleway_k8s_cluster", "minimal" ]
}`
	}

	return fmt.Sprintf(`
resource "scaleway_k8s_pool" "default" {
    name = "default"
	cluster_id = "${scaleway_k8s_cluster.minimal.id}"
	node_type = "gp1_xs"
	autohealing = true
	autoscaling = true
	size = 1
	tags = [ "terraform-test", "scaleway_k8s_cluster", "default" ]
}
resource "scaleway_k8s_cluster" "minimal" {
    name = "K8SPoolBetaConfigMinimal"
	cni = "calico"
	version = "%s"
	tags = [ "terraform-test", "scaleway_k8s_cluster", "minimal" ]
}
%s`, version, pool)
}

func testAccCheckScalewayK8SPoolConfigWait(version string, otherPool bool, otherPoolSize int) string {
	pool := ""
	if otherPool {
		pool += fmt.Sprintf(`
resource "scaleway_k8s_pool" "minimal" {
    name = "minimal"
	cluster_id = scaleway_k8s_cluster.minimal.id
	node_type = "gp1_xs"
	size = %d
	min_size = 1
	max_size = %d

	wait_for_pool_ready = true
}`, otherPoolSize, otherPoolSize)
	}

	return fmt.Sprintf(`
resource "scaleway_k8s_pool" "default" {
    name = "default"
	cluster_id = scaleway_k8s_cluster.minimal.id
	node_type = "gp1_xs"
	size = 1
	min_size = 1
	max_size = 1
	wait_for_pool_ready = true
}

resource "scaleway_k8s_cluster" "minimal" {
    name = "PoolBetaConfigWait"
	cni = "calico"
	version = "%s"
	tags = [ "terraform-test", "scaleway_k8s_cluster", "minimal" ]
}
%s`, version, pool)
}

func testAccCheckScalewayK8SPoolConfigPlacementGroup(version string) string {
	return fmt.Sprintf(`
resource "scaleway_instance_placement_group" "placement_group" {
  name        = "pool-placement-group"
  policy_type = "max_availability"
  policy_mode = "optional"
}

resource "scaleway_k8s_pool" "placement_group" {
    name = "placement_group"
	cluster_id = scaleway_k8s_cluster.placement_group.id
	node_type = "gp1_xs"
	placement_group_id = scaleway_instance_placement_group.placement_group.id
	size = 1
}

resource "scaleway_k8s_cluster" "placement_group" {
    name = "placement_group"
	cni = "calico"
	version = "%s"
	tags = [ "terraform-test", "scaleway_k8s_cluster", "placement_group" ]
}`, version)
}
