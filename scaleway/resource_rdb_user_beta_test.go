package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_user_beta", &resource.Sweeper{
		Name: "scaleway_rdb_user_beta",
		F:    testSweepRDBInstance,
	})
}

func TestAccScalewayRdbUserBeta(t *testing.T) {
	resourceName := "scaleway_rdb_user_beta.db_user"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayRdbInstanceBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayRdbUserConfig(rName, "titi", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbUserBetaExists("scaleway_rdb_instance_beta.main", resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "titi"),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "true"),
				),
			},
			{
				Config: testAccScalewayRdbUserConfig(rName, "tata", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbUserBetaExists("scaleway_rdb_instance_beta.main", resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "tata"),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "false"),
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

		rdbAPI, region, _, err := rdbAPIWithRegionAndID(testAccProvider.Meta(), instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		_, instanceId, userName, err := resourceScalewayRdbUserBetaParseId(userResource.Primary.ID)
		if err != nil {
			return err
		}

		users, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
			InstanceID: instanceId,
			Region:     region,
			Name:       &userName,
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

func testAccScalewayRdbUserConfigBase(rName string) string {
	return fmt.Sprintf(`
resource scaleway_rdb_instance_beta main {
    name = %[1]q
    node_type = "db-dev-s"
    engine = "PostgreSQL-12"
    is_ha_cluster = false
    tags = [ "terraform-test", "scaleway_rdb_user_beta", "minimal" ]
}`, rName)
}

func testAccScalewayRdbUserConfig(rName, userName, isAdmin string) string {
	return testAccScalewayRdbUserConfigBase(rName) + fmt.Sprintf(`
resource scaleway_rdb_user_beta db_user {
  instance_id = scaleway_rdb_instance_beta.main.id
  name = %[1]q
  password = "R34lP4sSw#Rd"
  is_admin = %[2]q
}`, userName, isAdmin)
}
