package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_instance", &resource.Sweeper{
		Name: "scaleway_rdb_instance",
		F:    testSweepRDBInstance,
	})
}

func testSweepRDBInstance(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		rdbAPI := rdb.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the rdb instance in (%s)", region)
		listInstances, err := rdbAPI.ListInstances(&rdb.ListInstancesRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing rdb instances in (%s) in sweeper: %s", region, err)
		}

		for _, instance := range listInstances.Instances {
			_, err := rdbAPI.DeleteInstance(&rdb.DeleteInstanceRequest{
				InstanceID: instance.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting rdb instance in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayRdbInstance_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_rdb_instance main {
						name = "test-rdb"
						node_type = "db-dev-s"
						engine = "PostgreSQL-11"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_instance", "minimal" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRdbExists(tt, "scaleway_rdb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "name", "test-rdb"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "node_type", "db-dev-s"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "engine", "PostgreSQL-11"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "is_ha_cluster", "false"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "disable_backup", "true"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "tags.1", "scaleway_rdb_instance"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "tags.2", "minimal"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_instance.main", "endpoint_ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_instance.main", "endpoint_port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_instance.main", "certificate"),
				),
			},
			{
				Config: `
					resource scaleway_rdb_instance main {
						name = "test-rdb"
						node_type = "db-dev-m"
						engine = "PostgreSQL-11"
						is_ha_cluster = true
						disable_backup = false
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s8cret"
						tags = [ "terraform-test", "scaleway_rdb_instance", "minimal" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRdbExists(tt, "scaleway_rdb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "name", "test-rdb"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "node_type", "db-dev-m"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "engine", "PostgreSQL-11"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "is_ha_cluster", "true"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "disable_backup", "false"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "password", "thiZ_is_v&ry_s8cret"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "tags.1", "scaleway_rdb_instance"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "tags.2", "minimal"),
				),
			},
		},
	})
}

func testAccCheckScalewayRdbExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		rdbAPI, region, ID, err := rdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = rdbAPI.GetInstance(&rdb.GetInstanceRequest{
			InstanceID: ID,
			Region:     region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayRdbInstanceDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_rdb_instance" {
				continue
			}

			rdbAPI, region, ID, err := rdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = rdbAPI.GetInstance(&rdb.GetInstanceRequest{
				InstanceID: ID,
				Region:     region,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("instance (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
