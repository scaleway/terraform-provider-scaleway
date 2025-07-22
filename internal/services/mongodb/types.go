package mongodb

import (
	"strings"

	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
)

func flattenPublicNetwork(endpoints []*mongodb.Endpoint) (any, bool) {
	publicFlat := []map[string]any(nil)

	for _, endpoint := range endpoints {
		if endpoint.PublicNetwork == nil {
			continue
		}

		publicFlat = append(publicFlat, map[string]any{
			"id":         endpoint.ID,
			"port":       endpoint.Port,
			"dns_record": endpoint.DNSRecord,
		})

		break
	}

	return publicFlat, len(publicFlat) != 0
}

func flattenPrivateNetwork(endpoints []*mongodb.Endpoint) (any, bool) {
	privateFlat := []map[string]any(nil)

	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork == nil {
			continue
		}

		privateFlat = append(privateFlat, map[string]any{
			"pn_id":       endpoint.PrivateNetwork.PrivateNetworkID,
			"id":          endpoint.ID,
			"port":        endpoint.Port,
			"dns_records": []string{endpoint.DNSRecord},
		})

		break
	}

	return privateFlat, len(privateFlat) != 0
}

func NormalizeMongoDBVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}

	return version
}
