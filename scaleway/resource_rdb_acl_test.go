package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_user", &resource.Sweeper{
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
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource scaleway_rdb_acl main {
						instance_id = scaleway_rdb_instance.main.id
						acl_rules = "1.2.3.4"
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbACLExists(tt, "scaleway_rdb_instance.main", "1.2.3.4"),
					resource.TestCheckResourceAttr("scaleway_rdb_acl.main", "acl_rules.0", "1.2.3.4"),
				),
			},
		},
	})
}

func testAccCheckRdbACLExists(tt *TestTools, instance string, acl string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceResource, ok := state.RootModule().Resources[instance]
		if !ok {
			return fmt.Errorf("resource not found: %s", instance)
		}

		userResource, ok := state.RootModule().Resources[acl]
		if !ok {
			return fmt.Errorf("resource not found: %s", acl)
		}

		rdbAPI, region, _, err := rdbAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		instanceID, err := resourceScalewayRdbACLParseID(userResource.Primary.ID)
		if err != nil {
			return err
		}

		aclRules, err := rdbAPI.ListInstanceACLRules(&rdb.ListInstanceACLRulesRequest{
			Region:     region,
			InstanceID: instanceID,
		})
		if err != nil {
			return err
		}

		if len(aclRules.Rules) != 1 {
			return fmt.Errorf("no acl rules found")
		}

		return nil
	}
}
