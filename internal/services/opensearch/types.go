package opensearch

import (
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
)

func flattenEndpoints(endpoints []*searchdbapi.Endpoint) []map[string]any {
	if len(endpoints) == 0 {
		return nil
	}

	result := make([]map[string]any, 0, len(endpoints))

	for _, endpoint := range endpoints {
		endpointMap := map[string]any{
			"id": endpoint.ID,
		}

		if len(endpoint.Services) > 0 {
			services := make([]map[string]any, 0, len(endpoint.Services))
			for _, service := range endpoint.Services {
				services = append(services, map[string]any{
					"name": service.Name,
					"port": int(service.Port),
					"url":  service.URL,
				})
			}

			endpointMap["services"] = services
		}

		if endpoint.Public != nil {
			endpointMap["public"] = true
		}

		if endpoint.PrivateNetwork != nil {
			endpointMap["public"] = false
			endpointMap["private_network_id"] = endpoint.PrivateNetwork.PrivateNetworkID
		}

		result = append(result, endpointMap)
	}

	return result
}
