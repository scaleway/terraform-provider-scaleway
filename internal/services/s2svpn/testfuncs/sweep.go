package s2svpntestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_s2s_vpn_connection", &resource.Sweeper{
		Name: "scaleway_s2s_vpn_connection",
		F:    testSweepConnection,
	})

	resource.AddTestSweepers("scaleway_s2s_vpn_gateway", &resource.Sweeper{
		Name:         "scaleway_s2s_vpn_gateway",
		F:            testSweepVPNGateway,
		Dependencies: []string{"scaleway_s2s_vpn_connection"},
	})

	resource.AddTestSweepers("scaleway_s2s_vpn_customer_gateway", &resource.Sweeper{
		Name:         "scaleway_s2s_vpn_customer_gateway",
		F:            testSweepCustomerGateway,
		Dependencies: []string{"scaleway_s2s_vpn_connection"},
	})

	resource.AddTestSweepers("scaleway_s2s_vpn_routing_policy", &resource.Sweeper{
		Name:         "scaleway_s2s_vpn_routing_policy",
		F:            testSweepRoutingPolicy,
		Dependencies: []string{"scaleway_s2s_vpn_connection"},
	})
}

func testSweepConnection(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		s2svpnAPI := s2s_vpn.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the s2s vpn connection in (%s)", region)

		listConnections, err := s2svpnAPI.ListConnections(&s2s_vpn.ListConnectionsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing s2s vpn connections in (%s) in sweeper: %w", region, err)
		}

		for _, connection := range listConnections.Connections {
			err := s2svpnAPI.DeleteConnection(&s2s_vpn.DeleteConnectionRequest{
				Region:       region,
				ConnectionID: connection.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting s2s vpn connection in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepVPNGateway(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		s2svpnAPI := s2s_vpn.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the s2s vpn gateway in (%s)", region)

		listGateways, err := s2svpnAPI.ListVpnGateways(&s2s_vpn.ListVpnGatewaysRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing s2s vpn gateways in (%s) in sweeper: %w", region, err)
		}

		for _, gateway := range listGateways.Gateways {
			_, err := s2svpnAPI.DeleteVpnGateway(&s2s_vpn.DeleteVpnGatewayRequest{
				Region:    region,
				GatewayID: gateway.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting s2s vpn gateway in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepCustomerGateway(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		s2svpnAPI := s2s_vpn.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the s2s vpn customer gateway in (%s)", region)

		listGateways, err := s2svpnAPI.ListCustomerGateways(&s2s_vpn.ListCustomerGatewaysRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing s2s vpn customer gateways in (%s) in sweeper: %w", region, err)
		}

		for _, gateway := range listGateways.Gateways {
			err := s2svpnAPI.DeleteCustomerGateway(&s2s_vpn.DeleteCustomerGatewayRequest{
				Region:    region,
				GatewayID: gateway.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting s2s vpn customer gateway in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepRoutingPolicy(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		s2svpnAPI := s2s_vpn.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the s2s vpn routing policy in (%s)", region)

		listPolicies, err := s2svpnAPI.ListRoutingPolicies(&s2s_vpn.ListRoutingPoliciesRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing s2s vpn routing policies in (%s) in sweeper: %w", region, err)
		}

		for _, policy := range listPolicies.RoutingPolicies {
			err := s2svpnAPI.DeleteRoutingPolicy(&s2s_vpn.DeleteRoutingPolicyRequest{
				Region:          region,
				RoutingPolicyID: policy.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting s2s vpn routing policy in sweeper: %w", err)
			}
		}

		return nil
	})
}
