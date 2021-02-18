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
	instanceName := "TestAccScalewayRdbACL_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
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
								ip = "1.2.3.4"
								description = "foo"
							}

						acl_rules {
								ip = "4.5.6.7"
								description = "bar"
							}
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0.description", "foo"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.ip", "4.5.6.7"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.1.description", "bar"),
				),
			},
		},
	})
}

//func testAccCheckRdbACLExists(tt *TestTools, instance string, acl string) resource.TestCheckFunc {
//	return func(state *terraform.State) error {
//		instanceResource, ok := state.RootModule().Resources[instance]
//		if !ok {
//			return fmt.Errorf("resource not found: %s", instance)
//		}
//
//		userResource, ok := state.RootModule().Resources[acl]
//		if !ok {
//			return fmt.Errorf("resource not found: %s", acl)
//		}
//
//		rdbAPI, region, _, err := rdbAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
//		if err != nil {
//			return err
//		}
//
//		instanceID, err := resourceScalewayRdbACLParseID(userResource.Primary.ID)
//		if err != nil {
//			return err
//		}
//
//		aclRules, err := rdbAPI.ListInstanceACLRules(&rdb.ListInstanceACLRulesRequest{
//			Region:     region,
//			InstanceID: instanceID,
//		})
//		if err != nil {
//			return err
//		}
//
//		if len(aclRules.Rules) != 1 {
//			return fmt.Errorf("no acl rules found")
//		}
//
//		return nil
//	}
//}
