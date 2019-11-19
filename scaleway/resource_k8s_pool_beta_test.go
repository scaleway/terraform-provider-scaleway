package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta3"
)

func TestAccScalewayK8SClusterPoolMinimal(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SPoolBetaConfigMinimal("1.16.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.minimal"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolBetaConfigMinimal("1.16.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.minimal"),
					testAccCheckScalewayK8SPoolBetaExists("scaleway_k8s_pool_beta.minimal"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool_beta.minimal", "node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool_beta.minimal", "size", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool_beta.minimal", "autohealing", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool_beta.minimal", "autoscaling", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool_beta.minimal", "version", "1.16.0"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_pool_beta.minimal", "id"),
				),
			},
			{
				Config: testAccCheckScalewayK8SPoolBetaConfigMinimal("1.16.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.minimal"),
					testAccCheckScalewayK8SPoolBetaDestroy,
				),
			},
		},
	})
}

func TestAccScalewayK8SClusterPoolPlacementGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayK8SClusterBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayK8SPoolBetaConfigPlacementGroup("1.16.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayK8SClusterBetaExists("scaleway_k8s_cluster_beta.placement_group"),
					testAccCheckScalewayK8SPoolBetaExists("scaleway_k8s_pool_beta.placement_group"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool_beta.placement_group", "node_type", "gp1_xs"),
					resource.TestCheckResourceAttr("scaleway_k8s_pool_beta.placement_group", "size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_pool_beta.placement_group", "id"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_pool_beta.placement_group", "placement_group_id"),
				),
			},
		},
	})
}

func testAccCheckScalewayK8SPoolBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_k8s_pool_beta" {
			continue
		}

		k8sAPI, region, poolID, err := getK8SAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = k8sAPI.GetPool(&k8s.GetPoolRequest{
			Region: region,
			PoolID: poolID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("Pool (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}
	return nil
}

func testAccCheckScalewayK8SPoolBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		k8sAPI, region, poolID, err := getK8SAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
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

func testAccCheckScalewayK8SPoolBetaConfigMinimal(version string, otherPool bool) string {
	pool := ""
	if otherPool {
		pool += fmt.Sprintf(`
resource "scaleway_k8s_pool_beta" "minimal" {
    name = "minimal"
	cluster_id = "${scaleway_k8s_cluster_beta.minimal.id}"
	node_type = "gp1_xs"
	autohealing = true
	autoscaling = true
	size = 1
}`)
	}

	return fmt.Sprintf(`
resource "scaleway_k8s_cluster_beta" "minimal" {
    name = "minimal"
	cni = "calico"
	version = "%s"
	default_pool {
		node_type = "gp1_xs"
		size = 1
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "minimal" ]
}
%s`, version, pool)
}

func testAccCheckScalewayK8SPoolBetaConfigPlacementGroup(version string) string {
	return fmt.Sprintf(`
resource "scaleway_instance_placement_group" "placement_group" {
  name        = "pool-placement-group"
  policy_type = "max_availability"
  policy_mode = "optional"
}

resource "scaleway_k8s_pool_beta" "placement_group" {
    name = "placement_group"
	cluster_id = scaleway_k8s_cluster_beta.placement_group.id
	node_type = "gp1_xs"
	placement_group_id = scaleway_instance_placement_group.placement_group.id
	size = 1
}

resource "scaleway_k8s_cluster_beta" "placement_group" {
    name = "placement_group"
	cni = "calico"
	version = "%s"
	default_pool {
		node_type = "gp1_xs"
		size = 1
	}
	tags = [ "terraform-test", "scaleway_k8s_cluster_beta", "placement_group" ]
}`, version)
}
