package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_apple_silicon_instance", &resource.Sweeper{
		Name: "scaleway_apple_silicon",
		F:    testSweepAppleSiliconServer,
	})
}

func testSweepAppleSiliconServer(_ string) error {
	return sweepZones([]scw.Zone{scw.ZoneFrPar1}, func(scwClient *scw.Client, zone scw.Zone) error {
		asAPI := applesilicon.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the apple silicon instance in (%s)", zone)
		listServers, err := asAPI.ListServers(&applesilicon.ListServersRequest{Zone: zone}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing apple silicon servers in (%s) in sweeper: %s", zone, err)
		}

		for _, server := range listServers.Servers {
			errDelete := asAPI.DeleteServer(&applesilicon.DeleteServerRequest{
				ServerID: server.ID,
				Zone:     zone,
			})
			if errDelete != nil {
				return fmt.Errorf("error deleting apple silicon server in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayAppleSiliconServer_Basic(t *testing.T) {
	t.Skip("Skipping AppleSilicon test as this kind of server can't be deleted before 24h")
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayAppleSiliconServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "test-m1"
						type = "M1-M"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAppleSiliconExists(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "name", "test-m1"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_server.main", "type", AppleSiliconM1Type),
					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "ip"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "vnc_url"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_server.main", "deletable_at"),
				),
			},
		},
	})
}

func testAccCheckScalewayAppleSiliconExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		asAPI, zone, ID, err := asAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = asAPI.GetServer(&applesilicon.GetServerRequest{
			ServerID: ID,
			Zone:     zone,
		})

		if err != nil {
			return fmt.Errorf("server not found: %s", err)
		}

		return nil
	}
}

func testAccCheckScalewayAppleSiliconServerDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_apple_silicon_server" {
				continue
			}

			asAPI, zone, ID, err := asAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = asAPI.GetServer(&applesilicon.GetServerRequest{
				ServerID: ID,
				Zone:     zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("server (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return fmt.Errorf("unexpected error when getting server (%s): %s", rs.Primary.ID, err)
			}
		}

		return nil
	}
}
