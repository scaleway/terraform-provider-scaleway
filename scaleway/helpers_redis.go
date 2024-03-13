package scaleway

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultRedisClusterTimeout           = 15 * time.Minute
	defaultWaitRedisClusterRetryInterval = 5 * time.Second
)

// newRedisApi returns a new Redis API
func newRedisAPI(m interface{}) *redis.API {
	return redis.NewAPI(meta.ExtractScwClient(m))
}

// redisAPIWithZone returns a new Redis API and the zone for a Create request
func redisAPIWithZone(d *schema.ResourceData, m interface{}) (*redis.API, scw.Zone, error) {
	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}
	return newRedisAPI(m), zone, nil
}

// redisAPIWithZoneAndID returns a Redis API with zone and ID extracted from the state
func redisAPIWithZoneAndID(m interface{}, id string) (*redis.API, scw.Zone, string, error) {
	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return newRedisAPI(m), zone, ID, nil
}

func waitForRedisCluster(ctx context.Context, api *redis.API, zone scw.Zone, id string, timeout time.Duration) (*redis.Cluster, error) {
	retryInterval := defaultWaitRedisClusterRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForCluster(&redis.WaitForClusterRequest{
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		ClusterID:     id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func expandRedisPrivateNetwork(data []interface{}) ([]*redis.EndpointSpec, error) {
	if data == nil {
		return nil, nil
	}
	epSpecs := make([]*redis.EndpointSpec, 0, len(data))

	for _, rawPN := range data {
		pn := rawPN.(map[string]interface{})
		pnID := locality.ExpandID(pn["id"].(string))
		rawIPs := pn["service_ips"].([]interface{})
		ips := []scw.IPNet(nil)
		spec := &redis.EndpointSpecPrivateNetworkSpec{
			ID: pnID,
		}
		if len(rawIPs) != 0 {
			for _, rawIP := range rawIPs {
				ip, err := types.ExpandIPNet(rawIP.(string))
				if err != nil {
					return epSpecs, err
				}
				ips = append(ips, ip)
			}
			spec.ServiceIPs = ips
		} else {
			spec.IpamConfig = &redis.EndpointSpecPrivateNetworkSpecIpamConfig{}
		}

		epSpecs = append(epSpecs, &redis.EndpointSpec{PrivateNetwork: spec})
	}
	return epSpecs, nil
}

func expandRedisACLSpecs(i interface{}) ([]*redis.ACLRuleSpec, error) {
	rules := []*redis.ACLRuleSpec(nil)

	for _, aclRule := range i.(*schema.Set).List() {
		rawRule := aclRule.(map[string]interface{})
		rule := &redis.ACLRuleSpec{}
		if ruleDescription, hasDescription := rawRule["description"]; hasDescription {
			rule.Description = ruleDescription.(string)
		}
		ip, err := types.ExpandIPNet(rawRule["ip"].(string))
		if err != nil {
			return nil, fmt.Errorf("failed to validate acl ip (%s): %w", rawRule["ip"].(string), err)
		}
		rule.IPCidr = ip
		rules = append(rules, rule)
	}

	return rules, nil
}

func flattenRedisACLs(aclRules []*redis.ACLRule) interface{} {
	flat := []map[string]interface{}(nil)
	for _, acl := range aclRules {
		flat = append(flat, map[string]interface{}{
			"id":          acl.ID,
			"ip":          acl.IPCidr.String(),
			"description": types.FlattenStringPtr(acl.Description),
		})
	}
	return flat
}

func expandRedisSettings(i interface{}) []*redis.ClusterSetting {
	rawSettings := i.(map[string]interface{})
	settings := []*redis.ClusterSetting(nil)
	for key, value := range rawSettings {
		settings = append(settings, &redis.ClusterSetting{
			Name:  key,
			Value: value.(string),
		})
	}
	return settings
}

func flattenRedisSettings(settings []*redis.ClusterSetting) interface{} {
	rawSettings := make(map[string]string)
	for _, setting := range settings {
		rawSettings[setting.Name] = setting.Value
	}
	return rawSettings
}

func flattenRedisPrivateNetwork(endpoints []*redis.Endpoint) (interface{}, bool) {
	pnFlat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork == nil {
			continue
		}
		pn := endpoint.PrivateNetwork
		fetchRegion, err := pn.Zone.Region()
		if err != nil {
			return diag.FromErr(err), false
		}
		pnRegionalID := regional.NewIDString(fetchRegion, pn.ID)
		serviceIps := []interface{}(nil)
		for _, ip := range pn.ServiceIPs {
			serviceIps = append(serviceIps, ip.String())
		}
		pnFlat = append(pnFlat, map[string]interface{}{
			"endpoint_id": endpoint.ID,
			"zone":        pn.Zone,
			"id":          pnRegionalID,
			"service_ips": serviceIps,
		})
	}
	return pnFlat, len(pnFlat) != 0
}

func flattenRedisPublicNetwork(endpoints []*redis.Endpoint) interface{} {
	pnFlat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PublicNetwork == nil {
			continue
		}
		ipsFlat := []interface{}(nil)
		for _, ip := range endpoint.IPs {
			ipsFlat = append(ipsFlat, ip.String())
		}
		pnFlat = append(pnFlat, map[string]interface{}{
			"id":   endpoint.ID,
			"port": int(endpoint.Port),
			"ips":  ipsFlat,
		})
		break
	}
	return pnFlat
}

func redisPrivateNetworkSetHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})
	if pnID, ok := m["id"]; ok {
		buf.WriteString(locality.ExpandID(pnID))
	}

	if serviceIPs, ok := m["service_ips"]; ok {
		// Sort the service IPs before generating the hash.
		ips := serviceIPs.([]interface{})
		sort.Slice(ips, func(i, j int) bool {
			return ips[i].(string) < ips[j].(string)
		})

		for i, item := range ips {
			buf.WriteString(fmt.Sprintf("%d-%s-", i, item.(string)))
		}
	}

	return StringHashcode(buf.String())
}
