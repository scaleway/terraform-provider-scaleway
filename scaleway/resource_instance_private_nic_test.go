package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_private_nic", &resource.Sweeper{
		Name: "scaleway_instance_private_nic",
		F:    testSweepInstancePrivateNic,
	})
}

func testSweepInstancePrivateNic(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		instanceAPI := instance.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the private nic in (%s)", zone)

		listPNResponse, err := instanceAPI.ListPrivateNICs(&instance.ListPrivateNICsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing private nic in sweeper: %s", err)
		}

		for _, privateNICs := range listPNResponse.PrivateNics {
			err := instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
				Zone:         zone,
				PrivateNicID: privateNICs.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting private nic in sweeper: %s", err)
			}

		}
		return nil
	})
}

func TestAccScalewayInstancePrivateNIC_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstancePrivateNICDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {}
					resource "scaleway_instance_server" "server01" {
						image = "f974feac-abae-4365-b988-8ec7d1cec10d"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server" ]
					}
					resource scaleway_instance_private_nic nic01 {
						server_id          = scaleway_instance_server.server01.id
						private_network_id = scaleway_vpc_private_network.pn01.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePrivateNICExists(
						tt,
						"scaleway_instance_private_nic.nic01",
					),
				),
			},
		},
	})
}

func testAccCheckScalewayInstancePrivateNICExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, innerID, outerID, err := instanceAPIWithZoneAndNestedID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetPrivateNIC(&instance.GetPrivateNICRequest{
			ServerID:     outerID,
			PrivateNicID: innerID,
			Zone:         zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayInstancePrivateNICDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_private_nic" {
				continue
			}

			instanceAPI, zone, innerID, outerID, err := instanceAPIWithZoneAndNestedID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetPrivateNIC(&instance.GetPrivateNICRequest{
				ServerID:     outerID,
				PrivateNicID: innerID,
				Zone:         zone,
			})

			if err == nil {
				return fmt.Errorf(
					"instance private NIC %s still exists",
					rs.Primary.ID,
				)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
