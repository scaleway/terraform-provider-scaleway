package datawarehouse

import (
	"time"

	datawarehouseapi "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// Helpers

const (
	defaultWaitRetryInterval = 30 * time.Second
)

func NewAPI(m interface{}) *datawarehouseapi.API {
	return datawarehouseapi.NewAPI(meta.ExtractScwClient(m))
}

func expandStringList(list []interface{}) []string {
	res := make([]string, len(list))
	for i, v := range list {
		res[i] = v.(string)
	}
	return res
}

func flattenStringList(list []string) []interface{} {
	res := make([]interface{}, len(list))
	for i, v := range list {
		res[i] = v
	}
	return res
}

func expandEndpoints(raw []interface{}) ([]*datawarehouseapi.EndpointSpec, error) {
	out := make([]*datawarehouseapi.EndpointSpec, 0, len(raw))
	for _, item := range raw {
		m := item.(map[string]interface{})
		spec := &datawarehouseapi.EndpointSpec{}
		if m["public"].(bool) {
			spec.Public = &datawarehouseapi.EndpointSpecPublicDetails{}
		}
		if v, ok := m["private_network_id"].(string); ok && v != "" {
			_, id, err := regional.ParseID(v)
			if err != nil {
				return nil, err
			}
			spec.PrivateNetwork = &datawarehouseapi.EndpointSpecPrivateNetworkDetails{PrivateNetworkID: id}
		}
		out = append(out, spec)
	}
	return out, nil
}

func flattenEndpoints(endpoints []*datawarehouseapi.Endpoint) []interface{} {
	out := make([]interface{}, 0, len(endpoints))
	for _, e := range endpoints {
		m := make(map[string]interface{})
		if e.Public != nil {
			m["public"] = true
		}
		if e.PrivateNetwork != nil {
			m["private_network_id"] = e.PrivateNetwork.PrivateNetworkID
		}
		out = append(out, m)
	}
	return out
}
