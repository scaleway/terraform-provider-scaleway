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

		// "DNSRecord" is a single string; Services is a slice—take the first service if present.
		protocol := ""
		port := 0

		if len(endpoint.Services) > 0 {
			protocol = string(endpoint.Services[0].Protocol)
			port = int(endpoint.Services[0].Port)
		}

		publicFlat = append(publicFlat, map[string]interface{}{
			"id":         endpoint.ID,
			"dns_record": endpoint.DNSRecord,
			"protocol":   protocol,
			"port":       port,
		})

		break
	}

	return publicFlat, len(publicFlat) != 0
}
