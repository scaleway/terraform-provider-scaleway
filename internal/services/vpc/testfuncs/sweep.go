package vpctestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	vpcSDK "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_vpc", &resource.Sweeper{
		Name:         "scaleway_vpc",
		F:            testSweepVPC,
		Dependencies: []string{"scaleway_vpc_private_network"},
	})

	resource.AddTestSweepers("scaleway_vpc_private_network", &resource.Sweeper{
		Name:         "scaleway_vpc_private_network",
		F:            testSweepVPCPrivateNetwork,
		Dependencies: []string{"scaleway_ipam_ip"},
	})
}

func testSweepVPC(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		vpcAPI := vpcSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting the VPCs in (%s)", region)

		listVPCs, err := vpcAPI.ListVPCs(&vpcSDK.ListVPCsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing secrets in (%s) in sweeper: %s", region, err)
		}

		for _, v := range listVPCs.Vpcs {
			if v.IsDefault {
				continue
			}
			err := vpcAPI.DeleteVPC(&vpcSDK.DeleteVPCRequest{
				VpcID:  v.ID,
				Region: region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting VPC in sweeper: %s", err)
			}
		}

		return nil
	})
}

func testSweepVPCPrivateNetwork(_ string) error {
	err := acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		vpcAPI := vpcSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the private network in (%s)", region)

		listPNResponse, err := vpcAPI.ListPrivateNetworks(&vpcSDK.ListPrivateNetworksRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing private network in sweeper: %s", err)
		}

		for _, pn := range listPNResponse.PrivateNetworks {
			err := vpcAPI.DeletePrivateNetwork(&vpcSDK.DeletePrivateNetworkRequest{
				Region:           region,
				PrivateNetworkID: pn.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting private network in sweeper: %s", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
