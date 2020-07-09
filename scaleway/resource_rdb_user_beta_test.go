package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
)

// TODO: refactor
func init() {
	resource.AddTestSweepers("scaleway_rdb_user_beta", &resource.Sweeper{
		Name: "scaleway_rdb_user_beta",
		F:    testSweepRDBInstance,
	})
}

func TestAccScalewayRdbUserBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckScalewayRdbInstanceBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource scaleway_rdb_instance_beta main {
    name = "test-terraform"
    node_type = "db-dev-s"
    engine = "PostgreSQL-12"
    is_ha_cluster = false
    user_name = "toto"
    password = "Tata#Titi42"
    tags = [ "terraform-test", "scaleway_rdb_user_beta", "minimal" ]
}

resource scaleway_rdb_user_beta db_user {
  instance_id = scaleway_rdb_instance_beta.main.id
  name = "titi"
  password = "R34lP4sSw#Rd"
  is_admin = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbUserBetaExists("scaleway_rdb_instance_beta.main", "scaleway_rdb_user_beta.db_user"),
					resource.TestCheckResourceAttr("scaleway_rdb_user_beta.db_user", "name", "titi"),
					resource.TestCheckResourceAttr("scaleway_rdb_user_beta.db_user", "is_admin", "true"),
				),
			},
		},
	})

}

func testAccCheckRdbUserBetaExists(instance string, user string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceResource, ok := s.RootModule().Resources[instance]
		if !ok {
			return fmt.Errorf("resource not found: %s", instance)
		}

		userResource, ok := s.RootModule().Resources[user]
		if !ok {
			return fmt.Errorf("resource not found: %s", user)
		}

		rdbAPI, region, instanceID, err := rdbAPIWithRegionAndID(testAccProvider.Meta(), instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		users, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
			InstanceID: instanceID,
			Region:     region,
			Name:       &userResource.Primary.ID,
		})

		if len(users.Users) != 1 {
			return fmt.Errorf("No user found")
		}

		if err != nil {
			return err
		}

		return nil
	}
}
