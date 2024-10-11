package mongodb

import (
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

func expandPrivateNetwork(data []interface{}) ([]*mongodb.EndpointSpec, error) {
	if data == nil {
		return nil, nil
	}
	epSpecs := make([]*mongodb.EndpointSpec, 0, len(data))
	for _, rawPN := range data {
		pn := rawPN.(map[string]interface{})
		pnID := locality.ExpandID(pn["id"].(string))

		spec := &mongodb.EndpointSpecPrivateNetworkDetails{
			PrivateNetworkID: pnID,
		}
		epSpecs = append(epSpecs, &mongodb.EndpointSpec{PrivateNetwork: spec})
	}
	return epSpecs, nil
}

func flattenPrivateNetwork(endpoints []*mongodb.Endpoint) (interface{}, bool) {
	pnFlat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork == nil {
			continue
		}
		pnFlat = append(pnFlat, map[string]interface{}{
			"endpoint_id": endpoint.ID,
			"id":          endpoint.PrivateNetwork.PrivateNetworkID,
			"ips":         endpoint.IPs,
			"port":        endpoint.Port,
			"dns_records": endpoint.DNSRecords,
		})
	}
	return pnFlat, len(pnFlat) != 0
}

func flattenPublicNetwork(endpoints []*mongodb.Endpoint) (interface{}, bool) {
	publicFlat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.Public == nil {
			continue
		}
		publicFlat = append(publicFlat, map[string]interface{}{
			"id":         endpoint.ID,
			"port":       endpoint.Port,
			"dns_record": endpoint.DNSRecords[0],
		})
		break
	}
	return publicFlat, len(publicFlat) != 0
}
