package k8s_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc" "main" {
						name = "TestAccACL_Basic"
					}

					resource "scaleway_vpc_private_network" "acl_basic" {
						vpc_id = scaleway_vpc.main.id
					}
			
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
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_basic", "acl_rules.*", map[string]string{
						"ip":              "1.2.3.4/32",
						"description":     "First rule",
						"scaleway_ranges": "false",
					}),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.0.id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc" "main" {
						name = "TestAccACL_Basic"
					}

					resource "scaleway_vpc_private_network" "acl_basic" {
						vpc_id = scaleway_vpc.main.id
					}

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
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_basic", "acl_rules.*", map[string]string{
						"ip":              "1.2.3.4/32",
						"description":     "",
						"scaleway_ranges": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_basic", "acl_rules.*", map[string]string{
						"ip":              "5.6.7.0/30",
						"description":     "",
						"scaleway_ranges": "false",
					}),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.1.id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc" "main" {
						name = "TestAccACL_Basic"
					}

					resource "scaleway_vpc_private_network" "acl_basic" {
						vpc_id = scaleway_vpc.main.id
					}

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
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_basic", "acl_rules.*", map[string]string{
						"ip":              "1.2.3.4/32",
						"description":     "First rule",
						"scaleway_ranges": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_basic", "acl_rules.*", map[string]string{
						"ip":              "",
						"description":     "Scaleway ranges rule",
						"scaleway_ranges": "true",
					}),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_k8s_acl.acl_basic", "acl_rules.1.id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc" "main" {
						name = "TestAccACL_Basic"
					}

					resource "scaleway_vpc_private_network" "acl_basic" {
						vpc_id = scaleway_vpc.main.id
					}

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
					resource "scaleway_vpc" "main" {
						name = "TestAccACL_Basic"
					}

					resource "scaleway_vpc_private_network" "acl_basic" {
						vpc_id = scaleway_vpc.main.id
					}

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

func TestAccACL_RulesOrder(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	clusterName := "k8s-acl-order"
	latestK8sVersion := testAccK8SClusterGetLatestK8SVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckK8SClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc main {}

					resource "scaleway_vpc_private_network" "acl_order" {
						vpc_id = scaleway_vpc.main.id
					}
			
					resource "scaleway_k8s_cluster" "acl_order" {
						name = "%s"
						version = "%s"
						cni = "cilium"
						delete_additional_resources = true
						private_network_id = scaleway_vpc_private_network.acl_order.id
					}
			
					resource "scaleway_k8s_acl" "acl_order" {
						cluster_id = scaleway_k8s_cluster.acl_order.id
						acl_rules {
							ip = "12.2.3.4/32"
							description = "First rule"
						}
						acl_rules {
							ip = "11.2.3.4/32"
							description = "Second rule"
						}
						acl_rules {
							ip = "1.2.3.7/32"
							description = "Third rule"
						}
						acl_rules {
							ip = "1.2.3.4/32"
							description = "Fourth rule"
						}
					}`, clusterName, latestK8sVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_k8s_acl.acl_order", "cluster_id", "scaleway_k8s_cluster.acl_order", "id"),
					resource.TestCheckResourceAttr("scaleway_k8s_acl.acl_order", "acl_rules.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_order", "acl_rules.*", map[string]string{
						"ip":          "12.2.3.4/32",
						"description": "First rule",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_order", "acl_rules.*", map[string]string{
						"ip":          "11.2.3.4/32",
						"description": "Second rule",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_order", "acl_rules.*", map[string]string{
						"ip":          "1.2.3.7/32",
						"description": "Third rule",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_k8s_acl.acl_order", "acl_rules.*", map[string]string{
						"ip":          "1.2.3.4/32",
						"description": "Fourth rule",
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc main {}

					resource "scaleway_vpc_private_network" "acl_order" {
						vpc_id = scaleway_vpc.main.id
					}
			
					resource "scaleway_k8s_cluster" "acl_order" {
						name = "%s"
						version = "%s"
						cni = "cilium"
						delete_additional_resources = true
						private_network_id = scaleway_vpc_private_network.acl_order.id
					}
			
					resource "scaleway_k8s_acl" "acl_order" {
						cluster_id = scaleway_k8s_cluster.acl_order.id
						acl_rules {
							ip = "12.2.3.4/32"
							description = "First rule"
						}
						acl_rules {
							ip = "11.2.3.4/32"
							description = "Second rule"
						}
						acl_rules {
							ip = "1.2.3.7/32"
							description = "Third rule"
						}
						acl_rules {
							ip = "1.2.3.4/32"
							description = "Fourth rule"
						}
					}`, clusterName, latestK8sVersion),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

		// Check if this is the case where ACL resource was deleted (expected = "0.0.0.0/0")
		if expected == "0.0.0.0/0" {
			// When ACL is deleted, we should have exactly 2 rules:
			// 1. One with IP = 0.0.0.0/0 and Description = "Automatically generated after scaleway_k8s_acl resource deletion"
			// 2. One with ScalewayRanges = true and Description = "Automatically generated after scaleway_k8s_acl resource deletion"
			if acls.TotalCount != 2 {
				return fmt.Errorf("expected 2 ACL rules after deletion, got %d rules: %+v", acls.TotalCount, acls.Rules)
			}

			// Check that we have exactly one rule with IP and one with ScalewayRanges
			hasIPRule := false
			hasScalewayRangesRule := false

			for _, rule := range acls.Rules {
				if rule.IP != nil && rule.IP.String() == "0.0.0.0/0" &&
					rule.Description == "Automatically generated after scaleway_k8s_acl resource deletion" {
					hasIPRule = true
				}

				if rule.ScalewayRanges != nil && *rule.ScalewayRanges &&
					rule.Description == "Automatically generated after scaleway_k8s_acl resource deletion" {
					hasScalewayRangesRule = true
				}
			}

			if !hasIPRule {
				return fmt.Errorf("expected rule with IP=0.0.0.0/0 after ACL deletion, but not found in: %+v", acls.Rules)
			}

			if !hasScalewayRangesRule {
				return fmt.Errorf("expected rule with ScalewayRanges=true after ACL deletion, but not found in: %+v", acls.Rules)
			}

			return nil
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
