package datawarehouse

import (
	datawarehouseapi "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

// flattenPublicNetwork scans through all endpoints and returns at most one "public" block.
// It returns ([]map[string]interface{}, true) if a public endpoint exists, or (nil, false) otherwise.
func flattenPublicNetwork(endpoints []*datawarehouseapi.Endpoint, region scw.Region) (any, bool) {
	publicFlat := make([]map[string]any, 0, 1)

	for _, endpoint := range endpoints {
		if endpoint.Public == nil {
			continue
		}

		services := make([]map[string]any, 0, len(endpoint.Services))

		for _, svc := range endpoint.Services {
			services = append(services, map[string]any{
				"protocol": string(svc.Protocol),
				"port":     int(svc.Port),
			})
		}

		publicFlat = append(publicFlat, map[string]any{
			"id":         regional.NewIDString(region, endpoint.ID),
			"dns_record": endpoint.DNSRecord,
			"services":   services,
		})

		break
	}

	return publicFlat, len(publicFlat) != 0
}

// flattenPrivateNetwork scans through all endpoints and returns at most one "private" block.
// It returns ([]map[string]interface{}, true) if a private endpoint exists, or (nil, false) otherwise.
func flattenPrivateNetwork(endpoints []*datawarehouseapi.Endpoint, region scw.Region) (any, bool) {
	privateFlat := make([]map[string]any, 0, 1)

	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork == nil {
			continue
		}

		services := make([]map[string]any, 0, len(endpoint.Services))

		for _, svc := range endpoint.Services {
			services = append(services, map[string]any{
				"protocol": string(svc.Protocol),
				"port":     int(svc.Port),
			})
		}

		privateFlat = append(privateFlat, map[string]any{
			"pn_id":      regional.NewIDString(region, endpoint.PrivateNetwork.PrivateNetworkID),
			"id":         regional.NewIDString(region, endpoint.ID),
			"dns_record": endpoint.DNSRecord,
			"services":   services,
		})

		break
	}

	return privateFlat, len(privateFlat) != 0
}
