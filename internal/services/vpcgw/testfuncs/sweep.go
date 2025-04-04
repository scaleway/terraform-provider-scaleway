package vpcgwtestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	vpcgwSDK "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_vpc_public_gateway_dhcp", &resource.Sweeper{
		Name: "scaleway_vpc_public_gateway_dhcp",
		F:    testSweepVPCPublicGatewayDHCP,
	})
	resource.AddTestSweepers("scaleway_vpc_public_gateway_ip", &resource.Sweeper{
		Name: "scaleway_vpc_public_gateway_ip",
		F:    testSweepVPCPublicGatewayIP,
	})
	resource.AddTestSweepers("scaleway_gateway_network", &resource.Sweeper{
		Name: "scaleway_gateway_network",
		F:    testSweepVPCGatewayNetwork,
	})
	resource.AddTestSweepers("scaleway_vpc_public_gateway", &resource.Sweeper{
		Name: "scaleway_vpc_public_gateway",
		F:    testSweepVPCPublicGateway,
	})
}

func testSweepVPCPublicGateway(_ string) error {
	return acctest.SweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		api := v2.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the public gateways in (%+v)", zone)

		listGatewayResponse, err := api.ListGateways(&v2.ListGatewaysRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing public gateway in sweeper: %w", err)
		}

		for _, gateway := range listGatewayResponse.Gateways {
			_, err := api.DeleteGateway(&v2.DeleteGatewayRequest{
				Zone:      zone,
				GatewayID: gateway.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting public gateway in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepVPCGatewayNetwork(_ string) error {
	return acctest.SweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		api := v2.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the gateway network in (%s)", zone)

		listPNResponse, err := api.ListGatewayNetworks(&v2.ListGatewayNetworksRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing gateway network in sweeper: %w", err)
		}

		for _, gn := range listPNResponse.GatewayNetworks {
			_, err := api.DeleteGatewayNetwork(&v2.DeleteGatewayNetworkRequest{
				GatewayNetworkID: gn.GatewayID,
				Zone:             zone,
			})
			if err != nil {
				return fmt.Errorf("error deleting gateway network in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepVPCPublicGatewayIP(_ string) error {
	return acctest.SweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		api := v2.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the public gateways ip in (%s)", zone)

		listIPResponse, err := api.ListIPs(&v2.ListIPsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing public gateway ip in sweeper: %w", err)
		}

		for _, ip := range listIPResponse.IPs {
			err := api.DeleteIP(&v2.DeleteIPRequest{
				Zone: zone,
				IPID: ip.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting public gateway ip in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepVPCPublicGatewayDHCP(_ string) error {
	return acctest.SweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		api := vpcgwSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying public gateway dhcps in (%+v)", zone)

		listDHCPsResponse, err := api.ListDHCPs(&vpcgwSDK.ListDHCPsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing public gateway dhcps in sweeper: %w", err)
		}

		for _, dhcp := range listDHCPsResponse.Dhcps {
			err := api.DeleteDHCP(&vpcgwSDK.DeleteDHCPRequest{
				Zone:   zone,
				DHCPID: dhcp.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting public gateway dhcp in sweeper: %w", err)
			}
		}

		return nil
	})
}
