package s2svpn

import (
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandVPNGatewayPublicConfig(raw any) *s2s_vpn.CreateVpnGatewayRequestPublicConfig {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &s2s_vpn.CreateVpnGatewayRequestPublicConfig{
		IpamIPv4ID: types.ExpandStringPtr(locality.ExpandID(rawMap["ipam_ipv4_id"].(string))),
		IpamIPv6ID: types.ExpandStringPtr(locality.ExpandID(rawMap["ipam_ipv6_id"].(string))),
	}
}

func expandConnectionRequestBgpConfig(raw any) (config *s2s_vpn.CreateConnectionRequestBgpConfig, err error) {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil, nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	privateIPNet, err := types.ExpandIPNet(rawMap["private_ip"].(string))
	if err != nil {
		return nil, err
	}

	peerPrivateIPNet, err := types.ExpandIPNet(rawMap["peer_private_ip"].(string))
	if err != nil {
		return nil, err
	}

	return &s2s_vpn.CreateConnectionRequestBgpConfig{
		RoutingPolicyID: locality.ExpandID(rawMap["routing_policy_id"].(string)),
		PrivateIP:       &privateIPNet,
		PeerPrivateIP:   &peerPrivateIPNet,
	}, nil
}

func expandPrefixFilters(raw any) ([]scw.IPNet, error) {
	if raw == nil {
		return nil, nil
	}

	rawList, ok := raw.([]any)
	if !ok || len(rawList) == 0 {
		return nil, nil
	}

	prefixes := make([]scw.IPNet, 0, len(rawList))
	for _, v := range rawList {
		ipNet, err := types.ExpandIPNet(v.(string))
		if err != nil {
			return nil, err
		}

		prefixes = append(prefixes, ipNet)
	}

	return prefixes, nil
}

func expandConnectionCiphers(raw any) []*s2s_vpn.ConnectionCipher {
	if raw == nil {
		return nil
	}

	rawList := raw.([]any)

	res := make([]*s2s_vpn.ConnectionCipher, 0, len(rawList))
	for _, item := range rawList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		c := &s2s_vpn.ConnectionCipher{
			Encryption: s2s_vpn.ConnectionEncryption(m["encryption"].(string)),
		}

		if v, ok := m["integrity"]; ok {
			val := s2s_vpn.ConnectionIntegrity(v.(string))
			c.Integrity = &val
		}

		if v, ok := m["dh_group"]; ok {
			val := s2s_vpn.ConnectionDhGroup(v.(string))
			c.DhGroup = &val
		}

		res = append(res, c)
	}

	return res
}

func flattenVPNGatewayPublicConfig(region scw.Region, config *s2s_vpn.VpnGatewayPublicConfig) any {
	if config == nil {
		return nil
	}

	return []map[string]any{
		{
			"ipam_ipv4_id": regional.NewIDString(region, types.FlattenStringPtr(config.IpamIPv4ID).(string)),
			"ipam_ipv6_id": regional.NewIDString(region, types.FlattenStringPtr(config.IpamIPv6ID).(string)),
		},
	}
}

func flattenBGPSession(region scw.Region, session *s2s_vpn.BgpSession) (any, error) {
	if session == nil {
		return nil, nil
	}

	privateIP, err := types.FlattenIPNet(session.PrivateIP)
	if err != nil {
		return nil, err
	}

	peerPrivateIP, err := types.FlattenIPNet(session.PeerPrivateIP)
	if err != nil {
		return nil, err
	}
	return []map[string]any{
		{
			"routing_policy_id": regional.NewIDString(region, session.RoutingPolicyID),
			"private_ip":        privateIP,
			"peer_private_ip":   peerPrivateIP,
		},
	}, nil
}

func FlattenPrefixFilters(prefixes []scw.IPNet) ([]string, error) {
	res := make([]string, 0, len(prefixes))

	for _, p := range prefixes {
		flattened, err := types.FlattenIPNet(p)
		if err != nil {
			return nil, err
		}

		res = append(res, flattened)
	}

	return res, nil
}

func flattenConnectionCiphers(ciphers []*s2s_vpn.ConnectionCipher) []any {
	if ciphers == nil || len(ciphers) == 0 {
		return nil
	}

	res := make([]any, 0, len(ciphers))

	for _, c := range ciphers {
		m := map[string]any{
			"encryption": c.Encryption.String(),
		}
		if c.Integrity != nil {
			m["integrity"] = c.Integrity.String()
		}
		if c.DhGroup != nil {
			m["dh_group"] = c.DhGroup.String()
		}
		res = append(res, m)
	}

	return res
}
