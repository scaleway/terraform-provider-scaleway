package k8s_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	k8sSDK "github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/k8s"
)

func TestAccACL_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	clusterName := "k8s-acl-basic"
	latestK8sVersion := testAccK8SClusterGetLatestK8SVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "acl_basic" {}
			
					resource "scaleway_k8s_cluster" "acl_basic" {
						name = "%s"
						version = "%s"
						cni = "cilium"
						delete_additional_resources = true
						private_network_id = scaleway_vpc_private_network.acl_basic.id
					}
			
					resource "scaleway_k8s_acl" "acl_basic" {
						cluster_id = scaleway_k8s_cluster.acl_basic.id
						acl_rules {
							ip = "1.2.3.4/32"
							description = "First rule"
						}
					}`, clusterName, latestK8sVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_k8s_acl.acl_basic", "cluster_id", "scaleway_k8s_cluster.acl_basic", "id"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "no_ip_allowed", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.#", "1"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.ip", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.scaleway_ranges", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.description", "First rule"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.0.id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "acl_basic" {}
			
					resource "scaleway_k8s_cluster" "acl_basic" {
						name = "%s"
						version = "%s"
						cni = "cilium"
						delete_additional_resources = true
						private_network_id = scaleway_vpc_private_network.acl_basic.id
					}
			
					resource "scaleway_k8s_acl" "acl_basic" {
						cluster_id = scaleway_k8s_cluster.acl_basic.id
						acl_rules {
							ip = "1.2.3.4/32"
						}
						acl_rules {
							ip = "5.6.7.0/30"
							scaleway_ranges = false
						}
					}`, clusterName, latestK8sVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_k8s_acl.acl_basic", "cluster_id", "scaleway_k8s_cluster.acl_basic", "id"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.#", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.ip", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.scaleway_ranges", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.description", ""),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.0.id"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.1.ip", "5.6.7.0/30"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.1.scaleway_ranges", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.1.description", ""),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.1.id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "acl_basic" {}
			
					resource "scaleway_k8s_cluster" "acl_basic" {
						name = "%s"
						version = "%s"
						cni = "cilium"
						delete_additional_resources = true
						private_network_id = scaleway_vpc_private_network.acl_basic.id
					}
			
					resource "scaleway_k8s_acl" "acl_basic" {
						cluster_id = scaleway_k8s_cluster.acl_basic.id
						acl_rules {
							ip = "1.2.3.4/32"
							description = "First rule"
						}
						acl_rules {
							scaleway_ranges = true
							description = "Scaleway ranges rule"
						}
					}`, clusterName, latestK8sVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_k8s_acl.acl_basic", "cluster_id", "scaleway_k8s_cluster.acl_basic", "id"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.#", "2"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.ip", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.scaleway_ranges", "false"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.0.description", "First rule"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.0.id"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.1.ip", ""),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.1.scaleway_ranges", "true"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.1.description", "Scaleway ranges rule"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.1.id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "acl_basic" {}
			
					resource "scaleway_k8s_cluster" "acl_basic" {
						name = "%s"
						version = "%s"
						cni = "cilium"
						delete_additional_resources = true
						private_network_id = scaleway_vpc_private_network.acl_basic.id
					}
			
					resource "scaleway_k8s_acl" "acl_basic" {
						cluster_id = scaleway_k8s_cluster.acl_basic.id
						no_ip_allowed = true
					}`, clusterName, latestK8sVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_k8s_acl.acl_basic", "cluster_id", "scaleway_k8s_cluster.acl_basic", "id"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_basic", "no_ip_allowed", "true"),
					resource.TestCheckNoResourceAttr("scaleway_k8s_acl.acl_basic", "acl_rules.#"),
					testAccCheckK8SClusterAllowedIPs(tt, "scaleway_k8s_cluster.acl_basic", ""),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "acl_basic" {}

					resource "scaleway_k8s_cluster" "acl_basic" {
						name = "%s"
						version = "%s"
						cni = "cilium"
						delete_additional_resources = true
						private_network_id = scaleway_vpc_private_network.acl_basic.id 
					}`, clusterName, latestK8sVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckK8SClusterAllowedIPs(tt, "scaleway_k8s_cluster.acl_basic", "0.0.0.0/0"),
				),
			},
		},
	})
}

func testAccCheckK8SClusterAllowedIPs(tt *acctest.TestTools, n string, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		k8sAPI, region, clusterID, err := k8s.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = k8sAPI.WaitForCluster(&k8sSDK.WaitForClusterRequest{
			Region:    region,
			ClusterID: clusterID,
		})
		if err != nil {
			return err
		}

		acls, err := k8sAPI.ListClusterACLRules(&k8sSDK.ListClusterACLRulesRequest{
			Region:    region,
			ClusterID: clusterID,
		})
		if err != nil {
			return err
		}

		switch {
		case expected == "" && acls.TotalCount == 0:
			return nil
		case expected != "" && acls.TotalCount == 1 && acls.Rules[0].IP != nil && acls.Rules[0].IP.String() == expected:
			return nil
		default:
			return fmt.Errorf("expected 1 ACL rule for subnet %q, got: %+v", expected, acls.Rules)
		}
	}
}
