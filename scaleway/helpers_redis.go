package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	redis "github.com/scaleway/scaleway-sdk-go/api/redis/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultRedisClusterTimeout           = 15 * time.Minute
	defaultWaitRedisClusterRetryInterval = 5 * time.Second
)

// newRedisApi returns a new Redis API
func newRedisAPI(m interface{}) *redis.API {
	meta := m.(*Meta)
	return redis.NewAPI(meta.scwClient)
}

// redisAPIWithZone returns a new Redis API and the zone for a Create request
func redisAPIWithZone(d *schema.ResourceData, m interface{}) (*redis.API, scw.Zone, error) {
	meta := m.(*Meta)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return newRedisAPI(m), zone, nil
}

// redisAPIWithZoneAndID returns a Redis API with zone and ID extracted from the state
func redisAPIWithZoneAndID(m interface{}, id string) (*redis.API, scw.Zone, string, error) {
	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return newRedisAPI(m), zone, ID, nil
}

func waitForRedisCluster(ctx context.Context, api *redis.API, zone scw.Zone, id string, timeout time.Duration) (*redis.Cluster, error) {
	retryInterval := defaultWaitRedisClusterRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	return api.WaitForCluster(&redis.WaitForClusterRequest{
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		ClusterID:     id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func expandRedisPrivateNetwork(data interface{}) ([]*redis.EndpointSpec, error) {
	if data == nil {
		return nil, nil
	}
	var epSpecs []*redis.EndpointSpec

	for _, rawPN := range data.([]interface{}) {
		pn := rawPN.(map[string]interface{})
		id := expandID(pn["id"].(string))
		rawIPs := pn["service_ips"].([]interface{})
		ips := []scw.IPNet(nil)
		for _, rawIP := range rawIPs {
			ip, err := expandIPNet(rawIP.(string))
			if err != nil {
				return epSpecs, err
			}
			ips = append(ips, ip)
		}
		spec := &redis.EndpointSpecPrivateNetworkSpec{
			ID:         id,
			ServiceIPs: ips,
		}
		epSpecs = append(epSpecs, &redis.EndpointSpec{PrivateNetwork: spec})
	}
	return epSpecs, nil
}

func flattenRedisPrivateNetwork(endpoints []*redis.Endpoint) (interface{}, bool) {
	pnFlat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork != nil {
			pn := endpoint.PrivateNetwork
			pnZonedID := newZonedIDString(pn.Zone, pn.ID)
			serviceIps := []interface{}(nil)
			for _, ip := range pn.ServiceIPs {
				serviceIps = append(serviceIps, ip.String())
			}
			pnFlat = append(pnFlat, map[string]interface{}{
				"id":                 endpoint.ID,
				"zone":               pn.Zone,
				"private_network_id": pnZonedID,
				"service_ips":        serviceIps,
			})
		}
	}
	return pnFlat, len(pnFlat) != 0
}

func flattenRedisPublicNetwork(endpoints []*redis.Endpoint) interface{} {
	pnFlat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PublicNetwork != nil {
			pnFlat = append(pnFlat, map[string]interface{}{
				"id":   endpoint.ID,
				"port": endpoint.Port,
				"ips":  endpoint.IPs,
			})
		}
	}
	return nil
func expandRedisACLSpecs(i interface{}) ([]*redis.ACLRuleSpec, error) {
	rules := []*redis.ACLRuleSpec(nil)

	for _, aclRule := range i.([]interface{}) {
		rawRule := aclRule.(map[string]interface{})
		rule := &redis.ACLRuleSpec{}
		if ruleDescription, hasDescription := rawRule["description"]; hasDescription {
			rule.Description = ruleDescription.(string)
		}
		ip, err := expandIPNet(rawRule["ip"].(string))
		if err != nil {
			return nil, fmt.Errorf("failed to validate acl ip (%s): %w", rawRule["ip"].(string), err)
		}
		rule.IP = ip
		rules = append(rules, rule)
	}

	return rules, nil
}

func flattenRedisACLs(aclRules []*redis.ACLRule) interface{} {
	flat := []map[string]interface{}(nil)
	for _, acl := range aclRules {
		flat = append(flat, map[string]interface{}{
			"id":          acl.ID,
			"ip":          acl.IP.String(),
			"description": flattenStringPtr(acl.Description),
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
