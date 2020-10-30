package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_user_beta", &resource.Sweeper{
		Name: "scaleway_rdb_user_beta",
		F:    testSweepRDBInstance,
	})
}

func TestAccScalewayRdbUser_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	instanceName := "TestAccScalewayRdbUser_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance_beta main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user_beta", "minimal" ]
					}

					resource scaleway_rdb_user_beta db_user {
						instance_id = scaleway_rdb_instance_beta.main.id
						name = "foo"
						password = "R34lP4sSw#Rd"
						is_admin = true
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbUserBetaExists(tt, "scaleway_rdb_instance_beta.main", "scaleway_rdb_user_beta.db_user"),
					resource.TestCheckResourceAttr("scaleway_rdb_user_beta.db_user", "name", "foo"),
					resource.TestCheckResourceAttr("scaleway_rdb_user_beta.db_user", "is_admin", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance_beta main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user_beta", "minimal" ]
					}

					resource scaleway_rdb_user_beta db_user {
						instance_id = scaleway_rdb_instance_beta.main.id
						name = "bar"
						password = "R34lP4sSw#Rd"
						is_admin = false
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbUserBetaExists(tt, "scaleway_rdb_instance_beta.main", "scaleway_rdb_user_beta.db_user"),
					resource.TestCheckResourceAttr("scaleway_rdb_user_beta.db_user", "name", "bar"),
					resource.TestCheckResourceAttr("scaleway_rdb_user_beta.db_user", "is_admin", "false"),
				),
			},
		},
	})
}

func testAccCheckRdbUserBetaExists(tt *TestTools, instance string, user string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceResource, ok := state.RootModule().Resources[instance]
		if !ok {
			return fmt.Errorf("resource not found: %s", instance)
		}

		userResource, ok := state.RootModule().Resources[user]
		if !ok {
			return fmt.Errorf("resource not found: %s", user)
		}

		rdbAPI, region, _, err := rdbAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		instanceID, userName, err := resourceScalewayRdbUserBetaParseID(userResource.Primary.ID)
		if err != nil {
			return err
		}

		users, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
			InstanceID: instanceID,
			Region:     region,
			Name:       &userName,
		})
		if err != nil {
			return err
		}

		if len(users.Users) != 1 {
			return fmt.Errorf("no user found")
		}

		return nil
	}
}
