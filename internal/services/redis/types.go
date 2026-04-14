package redis

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandPrivateNetwork(data []any) ([]*redis.EndpointSpec, error) {
	if data == nil {
		return nil, nil
	}

	epSpecs := make([]*redis.EndpointSpec, 0, len(data))

	for _, rawPN := range data {
		pn := rawPN.(map[string]any)
		pnID := locality.ExpandID(pn["id"].(string))
		rawIPs := pn["service_ips"].([]any)
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

func expandACLSpecs(i any) ([]*redis.ACLRuleSpec, error) {
	rules := []*redis.ACLRuleSpec(nil)

	for _, aclRule := range i.(*schema.Set).List() {
		rawRule := aclRule.(map[string]any)
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

func flattenACLs(aclRules []*redis.ACLRule) any {
	flat := make([]map[string]any, 0, len(aclRules))
	for _, acl := range aclRules {
		flat = append(flat, map[string]any{
			"id":          acl.ID,
			"ip":          acl.IPCidr.String(),
			"description": types.FlattenStringPtr(acl.Description),
		})
	}

	return flat
}

func expandSettings(i any) []*redis.ClusterSetting {
	rawSettings := i.(map[string]any)
	settings := make([]*redis.ClusterSetting, 0, len(rawSettings))

	for key, value := range rawSettings {
		settings = append(settings, &redis.ClusterSetting{
			Name:  key,
			Value: value.(string),
		})
	}

	return settings
}

func flattenSettings(settings []*redis.ClusterSetting) any {
	rawSettings := make(map[string]string)
	for _, setting := range settings {
		rawSettings[setting.Name] = setting.Value
	}

	return rawSettings
}

func flattenPrivateNetwork(endpoints []*redis.Endpoint) (any, bool) {
	pnFlat := []map[string]any(nil)

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

		serviceIps := []any(nil)
		for _, ip := range pn.ServiceIPs {
			serviceIps = append(serviceIps, ip.String())
		}

		ips := []any(nil)
		for _, ip := range endpoint.IPs {
			ips = append(ips, ip.String())
		}

		pnFlat = append(pnFlat, map[string]any{
			"endpoint_id": endpoint.ID,
			"zone":        pn.Zone,
			"id":          pnRegionalID,
			"port":        int(endpoint.Port),
			"ips":         ips,
			"service_ips": serviceIps,
		})
	}

	return pnFlat, len(pnFlat) != 0
}

func flattenPublicNetwork(endpoints []*redis.Endpoint) any {
	pnFlat := []map[string]any(nil)

	for _, endpoint := range endpoints {
		if endpoint.PublicNetwork == nil {
			continue
		}

		ipsFlat := make([]any, 0, len(endpoint.IPs))
		for _, ip := range endpoint.IPs {
			ipsFlat = append(ipsFlat, ip.String())
		}

		pnFlat = append(pnFlat, map[string]any{
			"id":   endpoint.ID,
			"port": int(endpoint.Port),
			"ips":  ipsFlat,
		})

		break
	}

	return pnFlat
}

// redisConnectionString builds a URI of the form redis(s)://[user:password@]host:port/0 for the first
// endpoint with a usable IP (public network preferred, then private).
// When password is set, user and password are set via url.UserPassword so clients send ACL AUTH (required e.g. on Redis 8).
// When password is empty (e.g. password_wo), userinfo is omitted because the secret is not in state.
func redisConnectionString(endpoints []*redis.Endpoint, username, password string, tlsEnabled bool) string {
	scheme := "redis"
	if tlsEnabled {
		scheme = "rediss"
	}

	host, port := redisConnectionHostPort(endpoints)
	if host == "" || port == 0 {
		return ""
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10)),
		Path:   "/0",
	}
	if password != "" {
		u.User = url.UserPassword(username, password)
	}

	return u.String()
}

func redisConnectionHostPort(endpoints []*redis.Endpoint) (host string, port uint32) {
	for _, endpoint := range endpoints {
		if endpoint == nil || endpoint.PublicNetwork == nil {
			continue
		}

		if len(endpoint.IPs) > 0 {
			return endpoint.IPs[0].String(), endpoint.Port
		}
	}

	for _, endpoint := range endpoints {
		if endpoint == nil || endpoint.PrivateNetwork == nil {
			continue
		}

		if len(endpoint.IPs) > 0 {
			return endpoint.IPs[0].String(), endpoint.Port
		}
	}

	return "", 0
}
