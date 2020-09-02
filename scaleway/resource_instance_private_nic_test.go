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

func testSweepInstancePrivateNic(zone string) error {
	z, err := scw.ParseZone(zone)
	if err != nil {
		return err
	}
	scwClient, err := sharedClientForZone(z)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	instanceAPI := instance.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the VPC private networks in %s", zone)
	listInstancePrivateNICs, err := instanceAPI.ListPrivateNICs(&instance.ListPrivateNICsRequest{}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing private networks in (%s) in sweeper: %s", zone, err)
	}

	for _, n := range listInstancePrivateNICs.PrivateNics {
		err := instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
			PrivateNicID: n.ID,
			Zone:         z,
		})
		if err != nil {
			return fmt.Errorf("error deleting private network in sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayInstancePrivateNIC(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstancePrivateNICDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_instance_private_nic nic01 {
						server_id: "4d67153f-24ee-4cdb-ae79-82986925b247"
						private_network_id: "7bad02dc-edfb-4235-ab37-8f57634dd1d1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePrivateNICExists("scaleway_instance_private_nic.nic01"),
					resource.TestCheckResourceAttr("scaleway_vpc_private_network.nic01", "server_id", "fr-par-1/4d67153f-24ee-4cdb-ae79-82986925b247"),
					resource.TestCheckResourceAttr("scaleway_vpc_private_network.nic01", "private_network_id", "fr-par-1/7bad02dc-edfb-4235-ab37-8f57634dd1d1"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstancePrivateNICExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, innerID, outerID, err := instanceAPIWithZoneAndNestedID(testAccProvider.Meta(), rs.Primary.ID)
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

func testAccCheckScalewayInstancePrivateNICDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_instance_private_nic" {
			continue
		}

		instanceAPI, zone, innerID, outerID, err := instanceAPIWithZoneAndNestedID(testAccProvider.Meta(), rs.Primary.ID)
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
				"Instance private NIC %s still exists",
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
