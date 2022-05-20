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

	cluster, err := api.WaitForCluster(&redis.WaitForClusterRequest{
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		ClusterID:     id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error waiting for redis cluster %q: %s", id, err)
	}

	return cluster, nil
}

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
