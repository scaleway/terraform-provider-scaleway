package kafka

import (
	kafkaapi "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
)

// flattenPublicNetwork scans through all endpoints and returns at most one "public" block.
// It returns ([]map[string]interface{}, true) if a public endpoint exists, or (nil, false) otherwise.
func flattenPublicNetwork(endpoints []*kafkaapi.Endpoint) (any, bool) {
	publicFlat := make([]map[string]any, 0, 1)

	for _, endpoint := range endpoints {
		if endpoint.PublicNetwork == nil {
			continue
		}

		dnsRecords := make([]any, 0, len(endpoint.DNSRecords))
		for _, dns := range endpoint.DNSRecords {
			dnsRecords = append(dnsRecords, dns)
		}

		publicFlat = append(publicFlat, map[string]any{
			"id":          endpoint.ID,
			"dns_records": dnsRecords,
			"port":        int(endpoint.Port),
		})

		break
	}

	return publicFlat, len(publicFlat) != 0
}

// flattenPrivateNetwork scans through all endpoints and returns at most one "private" block.
// It returns ([]map[string]interface{}, true) if a private endpoint exists, or (nil, false) otherwise.
func flattenPrivateNetwork(endpoints []*kafkaapi.Endpoint) (any, bool) {
	privateFlat := make([]map[string]any, 0, 1)

	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork == nil {
			continue
		}

		dnsRecords := make([]any, 0, len(endpoint.DNSRecords))
		for _, dns := range endpoint.DNSRecords {
			dnsRecords = append(dnsRecords, dns)
		}

		privateFlat = append(privateFlat, map[string]any{
			"pn_id":       endpoint.PrivateNetwork.PrivateNetworkID,
			"id":          endpoint.ID,
			"dns_records": dnsRecords,
			"port":        int(endpoint.Port),
		})

		break
	}

	return privateFlat, len(privateFlat) != 0
}
