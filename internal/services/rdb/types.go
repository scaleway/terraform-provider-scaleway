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

func flattenInstanceSettings(settings []*rdb.InstanceSetting) any {
	res := make(map[string]string)
	for _, value := range settings {
		res[value.Name] = value.Value
	}

	return res
}

func expandInstanceSettings(i any) []*rdb.InstanceSetting {
	rawRule := i.(map[string]any)
	res := make([]*rdb.InstanceSetting, 0, len(rawRule))

	for key, value := range rawRule {
		res = append(res, &rdb.InstanceSetting{
			Name:  key,
			Value: value.(string),
		})
	}

	return res
}

func expandPrivateNetwork(data any, exist bool, ipamConfig *bool, staticConfig *string) ([]*rdb.EndpointSpec, diag.Diagnostics) {
	if data == nil || !exist {
		return nil, nil
	}

	var diags diag.Diagnostics

	res := make([]*rdb.EndpointSpec, 0, len(data.([]any)))

	for _, pn := range data.([]any) {
		r := pn.(map[string]any)
		spec := &rdb.EndpointSpec{
			PrivateNetwork: &rdb.EndpointSpecPrivateNetwork{
				PrivateNetworkID: locality.ExpandID(r["pn_id"].(string)),
				IpamConfig:       &rdb.EndpointSpecPrivateNetworkIpamConfig{},
			},
		}

		if staticConfig != nil {
			// Normalize IP to CIDR notation if needed (e.g., 10.0.0.1 -> 10.0.0.1/32)
			normalizedIP, err := types.NormalizeIPToCIDR(*staticConfig)
			if err != nil {
				return nil, append(diags, diag.FromErr(fmt.Errorf("failed to normalize private_network ip_net (%s): %w", r["ip_net"], err))...)
			}

			ip, err := types.ExpandIPNet(normalizedIP)
			if err != nil {
				return nil, append(diags, diag.FromErr(fmt.Errorf("failed to parse private_network ip_net (%s): %w", normalizedIP, err))...)
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

func flattenPrivateNetwork(endpoints []*rdb.Endpoint) (any, bool) {
	pnI := []map[string]any(nil)

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

			enableIpam := false
			if endpoint.PrivateNetwork.ProvisioningMode == rdb.EndpointPrivateNetworkDetailsProvisioningModeIpam {
				enableIpam = true
			}

			pnI = append(pnI, map[string]any{
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

func flattenLoadBalancer(endpoints []*rdb.Endpoint) (any, bool) {
	flat := []map[string]any(nil)

	for _, endpoint := range endpoints {
		if endpoint.LoadBalancer != nil {
			flat = append(flat, map[string]any{
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

func expandReadReplicaEndpointsSpecDirectAccess(data any) *rdb.ReadReplicaEndpointSpec {
	if data == nil || len(data.([]any)) == 0 {
		return nil
	}

	return &rdb.ReadReplicaEndpointSpec{
		DirectAccess: new(rdb.ReadReplicaEndpointSpecDirectAccess),
	}
}

// expandReadReplicaEndpointsSpecPrivateNetwork expand read-replica private network endpoints from schema to specs
func expandReadReplicaEndpointsSpecPrivateNetwork(data any, ipamConfig *bool, staticConfig *string) (*rdb.ReadReplicaEndpointSpec, diag.Diagnostics) {
	if data == nil || len(data.([]any)) == 0 {
		return nil, nil
	}
	// private_network is a list of size 1
	data = data.([]any)[0]

	rawEndpoint := data.(map[string]any)

	var diags diag.Diagnostics

	endpoint := &rdb.ReadReplicaEndpointSpec{
		PrivateNetwork: &rdb.ReadReplicaEndpointSpecPrivateNetwork{
			PrivateNetworkID: locality.ExpandID(rawEndpoint["private_network_id"]),
			IpamConfig:       &rdb.ReadReplicaEndpointSpecPrivateNetworkIpamConfig{},
		},
	}

	if staticConfig != nil {
		// Normalize IP to CIDR notation if needed (e.g., 10.0.0.1 -> 10.0.0.1/32)
		normalizedIP, err := types.NormalizeIPToCIDR(*staticConfig)
		if err != nil {
			return nil, append(diags, diag.FromErr(fmt.Errorf("failed to normalize private_network service_ip (%s): %w", rawEndpoint["service_ip"], err))...)
		}

		ipNet, err := types.ExpandIPNet(normalizedIP)
		if err != nil {
			return nil, append(diags, diag.FromErr(fmt.Errorf("failed to parse private_network service_ip (%s): %w", normalizedIP, err))...)
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
func flattenReadReplicaEndpoints(endpoints []*rdb.Endpoint) (directAccess, privateNetwork any) {
	for _, endpoint := range endpoints {
		rawEndpoint := map[string]any{
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

			enableIpam := false
			if endpoint.PrivateNetwork.ProvisioningMode == rdb.EndpointPrivateNetworkDetailsProvisioningModeIpam {
				enableIpam = true
			}

			rawEndpoint["private_network_id"] = pnRegionalID
			rawEndpoint["service_ip"] = endpoint.PrivateNetwork.ServiceIP.String()
			rawEndpoint["zone"] = endpoint.PrivateNetwork.Zone
			rawEndpoint["enable_ipam"] = enableIpam
			privateNetwork = rawEndpoint
		}
	}

	// direct_access and private_network are lists

	if directAccess != nil {
		directAccess = []any{directAccess}
	}

	if privateNetwork != nil {
		privateNetwork = []any{privateNetwork}
	}

	return directAccess, privateNetwork
}

func expandInstanceLogsPolicy(i any) *rdb.LogsPolicy {
	policyConfigRaw := i.([]any)
	for _, policyRaw := range policyConfigRaw {
		policy := policyRaw.(map[string]any)

		return &rdb.LogsPolicy{
			MaxAgeRetention:    types.ExpandUint32Ptr(policy["max_age_retention"]),
			TotalDiskRetention: types.ExpandSize(policy["total_disk_retention"]),
		}
	}

	return nil
}

func flattenInstanceLogsPolicy(policy *rdb.LogsPolicy) any {
	p := []map[string]any{}
	if policy != nil {
		p = append(p, map[string]any{
			"max_age_retention":    types.FlattenUint32Ptr(policy.MaxAgeRetention),
			"total_disk_retention": types.FlattenSize(policy.TotalDiskRetention),
		})
	} else {
		return nil
	}

	return p
}
