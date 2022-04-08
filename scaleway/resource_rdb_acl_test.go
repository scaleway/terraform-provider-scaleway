package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_acl", &resource.Sweeper{
		Name: "scaleway_rdb_acl",
		F:    testSweepRDBInstance,
	})
}

func TestAccScalewayRdbACL_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	instanceName := "rdb-acl-basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_ip" "front1_ip" {
					}

					resource "scaleway_instance_ip" "front2_ip" {
					}

					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
					}

					resource "scaleway_rdb_acl" "main" {
					  instance_id = scaleway_rdb_instance.main.id

					  acl_rules {
						ip = "${scaleway_instance_ip.front1_ip.address}/32"
					  }

					  acl_rules {
						ip = "${scaleway_instance_ip.front2_ip.address}/32"
					  }
					}
				`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_rdb_acl.main", "acl_rules.0.description"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_acl.main", "acl_rules.1.description"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
					}

					resource scaleway_rdb_acl main {
						instance_id = scaleway_rdb_instance.main.id
						acl_rules {
							ip = "1.2.3.4/32"
							description = "foo"
						}

						acl_rules {
							ip = "4.5.6.7/32"
							description = "bar"
						}
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.ip", "4.5.6.7/32"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.description", "bar"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.ip", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.description", "foo"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
					}

					resource scaleway_rdb_acl main {
						instance_id = scaleway_rdb_instance.main.id
						acl_rules {
							ip = "1.2.3.4/32"
							description = "foo"
						}

						acl_rules {
							ip = "9.0.0.0/16"
							description = "baz"
						}
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.ip", "9.0.0.0/16"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.description", "baz"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.ip", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.description", "foo"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "name", "rdb-acl-basic"),
				),
			},
		},
	})
}
