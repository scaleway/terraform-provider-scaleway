package rdb

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenInstanceSettings(settings []*rdb.InstanceSetting) interface{} {
	res := make(map[string]string)
	for _, value := range settings {
		res[value.Name] = value.Value
	}

	return res
}

func expandInstanceSettings(i interface{}) []*rdb.InstanceSetting {
	rawRule := i.(map[string]interface{})
	res := make([]*rdb.InstanceSetting, 0, len(rawRule))
	for key, value := range rawRule {
		res = append(res, &rdb.InstanceSetting{
			Name:  key,
			Value: value.(string),
		})
	}

	return res
}

func expandPrivateNetwork(data interface{}, exist bool, ipamConfig *bool, staticConfig *string) ([]*rdb.EndpointSpec, diag.Diagnostics) {
	if data == nil || !exist {
		return nil, nil
	}
	var diags diag.Diagnostics

	res := make([]*rdb.EndpointSpec, 0, len(data.([]interface{})))
	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		spec := &rdb.EndpointSpec{
			PrivateNetwork: &rdb.EndpointSpecPrivateNetwork{
				PrivateNetworkID: locality.ExpandID(r["pn_id"].(string)),
				IpamConfig:       &rdb.EndpointSpecPrivateNetworkIpamConfig{},
			},
		}

		if staticConfig != nil {
			ip, err := types.ExpandIPNet(*staticConfig)
			if err != nil {
				return nil, append(diags, diag.FromErr(fmt.Errorf("failed to parse private_network ip_net (%s): %s", r["ip_net"], err))...)
			}
			spec.PrivateNetwork.ServiceIP = &ip
			spec.PrivateNetwork.IpamConfig = nil
			if ipamConfig != nil && *ipamConfig {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Detail:   "`ip_net` field is set so `enable_ipam` field will be ignored",
				})
			}
		} else if ipamConfig == nil || !*ipamConfig {
			return nil, diag.FromErr(errors.New("at least one of `ip_net` or `enable_ipam` (set to true) must be set"))
		}
		res = append(res, spec)
	}

	return res, diags
}

func expandLoadBalancer() *rdb.EndpointSpec {
	return &rdb.EndpointSpec{
		LoadBalancer: &rdb.EndpointSpecLoadBalancer{},
	}
}

func flattenPrivateNetwork(endpoints []*rdb.Endpoint, enableIpam bool) (interface{}, bool) {
	pnI := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork != nil {
			pn := endpoint.PrivateNetwork
			fetchRegion, err := pn.Zone.Region()
			if err != nil {
				return diag.FromErr(err), false
			}
			pnRegionalID := regional.NewIDString(fetchRegion, pn.PrivateNetworkID)
			serviceIP, err := types.FlattenIPNet(pn.ServiceIP)
			if err != nil {
				return pnI, false
			}
			pnI = append(pnI, map[string]interface{}{
				"endpoint_id": endpoint.ID,
				"ip":          types.FlattenIPPtr(endpoint.IP),
				"port":        int(endpoint.Port),
				"name":        endpoint.Name,
				"ip_net":      serviceIP,
				"pn_id":       pnRegionalID,
				"hostname":    types.FlattenStringPtr(endpoint.Hostname),
				"enable_ipam": enableIpam,
			})
			return pnI, true
		}
	}

	return pnI, false
}

func flattenLoadBalancer(endpoints []*rdb.Endpoint) (interface{}, bool) {
	flat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.LoadBalancer != nil {
			flat = append(flat, map[string]interface{}{
				"endpoint_id": endpoint.ID,
				"ip":          types.FlattenIPPtr(endpoint.IP),
				"port":        int(endpoint.Port),
				"name":        endpoint.Name,
				"hostname":    types.FlattenStringPtr(endpoint.Hostname),
			})
			return flat, true
		}
	}

	return flat, false
}

func expandReadReplicaEndpointsSpecDirectAccess(data interface{}) *rdb.ReadReplicaEndpointSpec {
	if data == nil || len(data.([]interface{})) == 0 {
		return nil
	}

	return &rdb.ReadReplicaEndpointSpec{
		DirectAccess: new(rdb.ReadReplicaEndpointSpecDirectAccess),
	}
}

// expandReadReplicaEndpointsSpecPrivateNetwork expand read-replica private network endpoints from schema to specs
func expandReadReplicaEndpointsSpecPrivateNetwork(data interface{}, ipamConfig *bool, staticConfig *string) (*rdb.ReadReplicaEndpointSpec, diag.Diagnostics) {
	if data == nil || len(data.([]interface{})) == 0 {
		return nil, nil
	}
	// private_network is a list of size 1
	data = data.([]interface{})[0]

	rawEndpoint := data.(map[string]interface{})
	var diags diag.Diagnostics

	endpoint := &rdb.ReadReplicaEndpointSpec{
		PrivateNetwork: &rdb.ReadReplicaEndpointSpecPrivateNetwork{
			PrivateNetworkID: locality.ExpandID(rawEndpoint["private_network_id"]),
			IpamConfig:       &rdb.ReadReplicaEndpointSpecPrivateNetworkIpamConfig{},
		},
	}

	if staticConfig != nil {
		ipNet, err := types.ExpandIPNet(*staticConfig)
		if err != nil {
			return nil, append(diags, diag.FromErr(fmt.Errorf("failed to parse private_network service_ip (%s): %s", rawEndpoint["service_ip"], err))...)
		}
		endpoint.PrivateNetwork.ServiceIP = &ipNet
		endpoint.PrivateNetwork.IpamConfig = nil
		if ipamConfig != nil && !*ipamConfig {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Detail:   "`service_ip` field is set so `enable_ipam` field will be ignored",
			})
		}
	} else if ipamConfig == nil || !*ipamConfig {
		return nil, diag.FromErr(errors.New("at least one of `service_ip` or `enable_ipam` (set to true) must be set"))
	}

	return endpoint, diags
}

// flattenReadReplicaEndpoints flatten read-replica endpoints to directAccess and privateNetwork
func flattenReadReplicaEndpoints(endpoints []*rdb.Endpoint, enableIpam bool) (directAccess, privateNetwork interface{}) {
	for _, endpoint := range endpoints {
		rawEndpoint := map[string]interface{}{
			"endpoint_id": endpoint.ID,
			"ip":          types.FlattenIPPtr(endpoint.IP),
			"port":        int(endpoint.Port),
			"name":        endpoint.Name,
			"hostname":    types.FlattenStringPtr(endpoint.Hostname),
		}
		if endpoint.DirectAccess != nil {
			directAccess = rawEndpoint
		}
		if endpoint.PrivateNetwork != nil {
			fetchRegion, err := endpoint.PrivateNetwork.Zone.Region()
			if err != nil {
				return diag.FromErr(err), false
			}
			pnRegionalID := regional.NewIDString(fetchRegion, endpoint.PrivateNetwork.PrivateNetworkID)
			rawEndpoint["private_network_id"] = pnRegionalID
			rawEndpoint["service_ip"] = endpoint.PrivateNetwork.ServiceIP.String()
			rawEndpoint["zone"] = endpoint.PrivateNetwork.Zone
			rawEndpoint["enable_ipam"] = enableIpam
			privateNetwork = rawEndpoint
		}
	}

	// direct_access and private_network are lists

	if directAccess != nil {
		directAccess = []interface{}{directAccess}
	}
	if privateNetwork != nil {
		privateNetwork = []interface{}{privateNetwork}
	}

	return directAccess, privateNetwork
}
