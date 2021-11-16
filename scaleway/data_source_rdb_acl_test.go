package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceRDBAcl_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	instanceName := "data-source-rdb-acl-basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
					}

					resource "scaleway_rdb_acl" "main" {
						instance_id = scaleway_rdb_instance.main.id
						acl_rules {
							ip = "1.2.3.4/32"
							description = "foo"
						}

						acl_rules {
							ip = "4.5.6.7/32"
							description = "bar"
						}
					}
					`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.ip", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.ip", "4.5.6.7/32"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
					}

					resource "scaleway_rdb_acl" "main" {
						instance_id = scaleway_rdb_instance.main.id
						acl_rules {
							ip = "1.2.3.4/32"
							description = "foo"
						}

						acl_rules {
							ip = "4.5.6.7/32"
							description = "bar"
						}
					}
					data "scaleway_rdb_acl" "maindata" {
						instance_id = scaleway_rdb_instance.main.id

					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.ip", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.ip", "4.5.6.7/32"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "name", "data-source-rdb-acl-basic"),
				),
			},
		},
	})
}
