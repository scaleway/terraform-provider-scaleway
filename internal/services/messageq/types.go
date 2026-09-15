package messageq

import (
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
)

func ExpandEndpointSpecsFromPrivateNetwork(pnID string) []*messageqapi.EndpointSpec {
	if pnID == "" {
		return []*messageqapi.EndpointSpec{
			{
				Public: &messageqapi.EndpointSpecPublicDetails{},
			},
		}
	}

	return []*messageqapi.EndpointSpec{
		{
			PrivateNetwork: &messageqapi.EndpointSpecPrivateNetworkDetails{
				PrivateNetworkID: pnID,
			},
		},
	}
}
