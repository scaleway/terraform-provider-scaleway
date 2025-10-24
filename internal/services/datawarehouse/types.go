package datawarehouse

import (
	datawarehouseapi "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
)

// flattenPublicNetwork scans through all endpoints and returns at most one "public" block.
// It returns ([]map[string]interface{}, true) if a public endpoint exists, or (nil, false) otherwise.
func flattenPublicNetwork(endpoints []*datawarehouseapi.Endpoint) (interface{}, bool) {
	publicFlat := make([]map[string]interface{}, 0, 1)

	for _, endpoint := range endpoints {
		// skip any endpoint that is not public
		if endpoint.Public == nil {
			continue
		}

		// Flatten all services
		services := make([]map[string]interface{}, 0, len(endpoint.Services))

		for _, svc := range endpoint.Services {
			services = append(services, map[string]interface{}{
				"protocol": string(svc.Protocol),
				"port":     int(svc.Port),
			})
		}

		publicFlat = append(publicFlat, map[string]interface{}{
			"id":         endpoint.ID,
			"dns_record": endpoint.DNSRecord,
			"services":   services,
		})

		break
	}

	return publicFlat, len(publicFlat) != 0
}
