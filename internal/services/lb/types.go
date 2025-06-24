package lb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenPrivateNetworkConfigs(privateNetworks []*lb.PrivateNetwork) any {
	if len(privateNetworks) == 0 || privateNetworks == nil {
		return nil
	}

	pnI := []map[string]any(nil)

	var dhcpConfigExist bool

	for _, pn := range privateNetworks {
		if pn.DHCPConfig != nil { //nolint:staticcheck
			dhcpConfigExist = true
		}

		pnRegion, err := pn.LB.Zone.Region()
		if err != nil {
			return diag.FromErr(err)
		}

		pnRegionalID := regional.NewIDString(pnRegion, pn.PrivateNetworkID)
		pnI = append(pnI, map[string]any{
			"private_network_id": pnRegionalID,
			"dhcp_config":        dhcpConfigExist,
			"status":             pn.Status.String(),
			"zone":               pn.LB.Zone.String(),
			"static_config":      flattenLbPrivateNetworkStaticConfig(pn.StaticConfig), //nolint:staticcheck
			"ipam_ids":           regional.NewRegionalIDs(pnRegion, pn.IpamIDs),
		})
	}

	return pnI
}

func flattenLbACLMatch(match *lb.ACLMatch) any {
	if match == nil {
		return nil
	}

	return []map[string]any{
		{
			"ip_subnet":          types.FlattenSliceStringPtr(match.IPSubnet),
			"http_filter":        match.HTTPFilter.String(),
			"http_filter_value":  types.FlattenSliceStringPtr(match.HTTPFilterValue),
			"http_filter_option": match.HTTPFilterOption,
			"invert":             match.Invert,
			"ips_edge_services":  match.IPsEdgeServices,
		},
	}
}

func isIPSubnetConfigured(d *schema.ResourceData) bool {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() {
		return false
	}

	matchConfig := rawConfig.GetAttr("match")
	if matchConfig.IsNull() || matchConfig.LengthInt() == 0 {
		return false
	}

	matchBlock := matchConfig.Index(cty.NumberIntVal(0))
	if matchBlock.IsNull() || !matchBlock.Type().HasAttribute("ip_subnet") {
		return false
	}

	return !matchBlock.GetAttr("ip_subnet").IsNull()
}

func expandLbACLMatch(d *schema.ResourceData, raw any) *lb.ACLMatch {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)
	ipsEdgeServices := rawMap["ips_edge_services"].(bool)
	ipSubnetConfigured := isIPSubnetConfigured(d)

	var ipSubnet []*string

	switch {
	case ipsEdgeServices:
		ipSubnet = nil
	case ipSubnetConfigured:
		ipSubnet = types.ExpandSliceStringPtr(rawMap["ip_subnet"].([]any))
	default:
		ipSubnet = []*string{types.ExpandStringPtr("0.0.0.0/0")}
	}

	return &lb.ACLMatch{
		IPSubnet:         ipSubnet,
		IPsEdgeServices:  rawMap["ips_edge_services"].(bool),
		HTTPFilter:       lb.ACLHTTPFilter(rawMap["http_filter"].(string)),
		HTTPFilterValue:  types.ExpandSliceStringPtr(rawMap["http_filter_value"].([]any)),
		HTTPFilterOption: types.ExpandStringPtr(rawMap["http_filter_option"].(string)),
		Invert:           rawMap["invert"].(bool),
	}
}

func expandLbBackendMarkdownAction(raw any) lb.OnMarkedDownAction {
	if raw == "none" {
		return lb.OnMarkedDownActionOnMarkedDownActionNone
	}

	return lb.OnMarkedDownAction(raw.(string))
}

func flattenLbProtocol(protocol lb.Protocol) any {
	return protocol.String()
}

func expandLbProtocol(raw any) lb.Protocol {
	return lb.Protocol(raw.(string))
}

func flattenLbForwardPortAlgorithm(algo lb.ForwardPortAlgorithm) any {
	return algo.String()
}

func expandLbForwardPortAlgorithm(raw any) lb.ForwardPortAlgorithm {
	return lb.ForwardPortAlgorithm(raw.(string))
}

func flattenLbStickySessionsType(t lb.StickySessionsType) any {
	return t.String()
}

func expandLbStickySessionsType(raw any) lb.StickySessionsType {
	return lb.StickySessionsType(raw.(string))
}

func flattenLbHCTCP(config *lb.HealthCheckTCPConfig) any {
	if config == nil {
		return nil
	}

	return []map[string]any{
		{},
	}
}

func expandLbHCTCP(raw any) *lb.HealthCheckTCPConfig {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	return &lb.HealthCheckTCPConfig{}
}

func flattenLbHCHTTP(config *lb.HealthCheckHTTPConfig) any {
	if config == nil {
		return nil
	}

	return []map[string]any{
		{
			"uri":         config.URI,
			"method":      config.Method,
			"code":        types.FlattenInt32Ptr(config.Code),
			"host_header": config.HostHeader,
		},
	}
}

func expandLbHCHTTP(raw any) *lb.HealthCheckHTTPConfig {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &lb.HealthCheckHTTPConfig{
		URI:        rawMap["uri"].(string),
		Method:     rawMap["method"].(string),
		Code:       types.ExpandInt32Ptr(rawMap["code"]),
		HostHeader: rawMap["host_header"].(string),
	}
}

func flattenLbHCHTTPS(config *lb.HealthCheckHTTPSConfig) any {
	if config == nil {
		return nil
	}

	return []map[string]any{
		{
			"uri":         config.URI,
			"method":      config.Method,
			"code":        types.FlattenInt32Ptr(config.Code),
			"host_header": config.HostHeader,
			"sni":         config.Sni,
		},
	}
}

func expandLbHCHTTPS(raw any) *lb.HealthCheckHTTPSConfig {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &lb.HealthCheckHTTPSConfig{
		URI:        rawMap["uri"].(string),
		Method:     rawMap["method"].(string),
		Code:       types.ExpandInt32Ptr(rawMap["code"]),
		HostHeader: rawMap["host_header"].(string),
		Sni:        rawMap["sni"].(string),
	}
}

func expandLbLetsEncrypt(raw any) *lb.CreateCertificateRequestLetsencryptConfig {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)
	alternativeNames := rawMap["subject_alternative_name"].([]any)
	config := &lb.CreateCertificateRequestLetsencryptConfig{
		CommonName: rawMap["common_name"].(string),
	}

	for _, alternativeName := range alternativeNames {
		config.SubjectAlternativeName = append(config.SubjectAlternativeName, alternativeName.(string))
	}

	return config
}

func expandLbCustomCertificate(raw any) *lb.CreateCertificateRequestCustomCertificate {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)
	config := &lb.CreateCertificateRequestCustomCertificate{
		CertificateChain: rawMap["certificate_chain"].(string),
	}

	return config
}

func expandLbProxyProtocol(raw any) lb.ProxyProtocol {
	return lb.ProxyProtocol("proxy_protocol_" + raw.(string))
}

func flattenLbProxyProtocol(pp lb.ProxyProtocol) any {
	return strings.TrimPrefix(pp.String(), "proxy_protocol_")
}

func flattenLbBackendMarkdownAction(action lb.OnMarkedDownAction) any {
	if action == lb.OnMarkedDownActionOnMarkedDownActionNone {
		return "none"
	}

	return action.String()
}

func flattenLbACL(acl *lb.ACL) any {
	res := map[string]any{
		"name":   acl.Name,
		"match":  flattenLbACLMatch(acl.Match),
		"action": flattenLbACLAction(acl.Action),
	}

	return res
}

// expandLbACL transforms a state acl to an api one.
func expandLbACL(d *schema.ResourceData, i any) *lb.ACL {
	rawRule := i.(map[string]any)
	acl := &lb.ACL{
		Name:        rawRule["name"].(string),
		Description: rawRule["description"].(string),
		Match:       expandLbACLMatch(d, rawRule["match"]),
		Action:      expandLbACLAction(rawRule["action"]),
		CreatedAt:   types.ExpandTimePtr(rawRule["created_at"]),
		UpdatedAt:   types.ExpandTimePtr(rawRule["updated_at"]),
	}

	if rawRule["index"] != nil {
		acl.Index = int32(rawRule["index"].(int))
	}
	// remove http filter values if we do not pass any http filter
	if acl.Match.HTTPFilter == "" || acl.Match.HTTPFilter == lb.ACLHTTPFilterACLHTTPFilterNone {
		acl.Match.HTTPFilter = lb.ACLHTTPFilterACLHTTPFilterNone
		acl.Match.HTTPFilterValue = []*string{}
	}

	return acl
}

func flattenLbACLAction(action *lb.ACLAction) any {
	return []map[string]any{
		{
			"type":     action.Type,
			"redirect": flattenLbACLActionRedirect(action.Redirect),
		},
	}
}

func expandLbACLAction(raw any) *lb.ACLAction {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &lb.ACLAction{
		Type:     lb.ACLActionType(rawMap["type"].(string)),
		Redirect: expandLbACLActionRedirect(rawMap["redirect"]),
	}
}

func flattenLbACLActionRedirect(redirect *lb.ACLActionRedirect) any {
	if redirect == nil {
		return nil
	}

	return []map[string]any{
		{
			"type":   redirect.Type,
			"target": redirect.Target,
			"code":   redirect.Code,
		},
	}
}

func expandLbACLActionRedirect(raw any) *lb.ACLActionRedirect {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &lb.ACLActionRedirect{
		Type:   lb.ACLActionRedirectRedirectType(rawMap["type"].(string)),
		Target: rawMap["target"].(string),
		Code:   types.ExpandInt32Ptr(rawMap["code"]),
	}
}

func expandPrivateNetworks(data any) ([]*lb.PrivateNetwork, error) {
	if data == nil {
		return nil, nil
	}

	pns := []*lb.PrivateNetwork(nil)

	for _, pn := range data.(*schema.Set).List() {
		rawPn := pn.(map[string]any)
		privateNetwork := &lb.PrivateNetwork{}
		privateNetwork.PrivateNetworkID = locality.ExpandID(rawPn["private_network_id"].(string))

		if staticConfig, hasStaticConfig := rawPn["static_config"]; hasStaticConfig && len(staticConfig.([]any)) > 0 {
			privateNetwork.StaticConfig = expandLbPrivateNetworkStaticConfig(staticConfig) //nolint:staticcheck
		} else {
			privateNetwork.DHCPConfig = expandLbPrivateNetworkDHCPConfig(rawPn["dhcp_config"]) //nolint:staticcheck
		}

		privateNetwork.IpamIDs = locality.ExpandIDs(rawPn["ipam_ids"])

		pns = append(pns, privateNetwork)
	}

	return pns, nil
}

func expandLbPrivateNetworkStaticConfig(raw any) *lb.PrivateNetworkStaticConfig {
	if raw == nil || len(raw.([]any)) < 1 {
		return nil
	}

	return &lb.PrivateNetworkStaticConfig{
		IPAddress: types.ExpandStringsPtr(raw),
	}
}

func flattenLbPrivateNetworkStaticConfig(cfg *lb.PrivateNetworkStaticConfig) []string {
	if cfg == nil {
		return nil
	}

	return *cfg.IPAddress //nolint:staticcheck
}

func expandLbPrivateNetworkDHCPConfig(raw any) *lb.PrivateNetworkDHCPConfig {
	if raw == nil || !raw.(bool) {
		return nil
	}

	return &lb.PrivateNetworkDHCPConfig{}
}

func attachLBPrivateNetworks(ctx context.Context, lbAPI *lb.ZonedAPI, zone scw.Zone, pnConfigs []*lb.PrivateNetwork, lbID string, timeout time.Duration) ([]*lb.PrivateNetwork, error) {
	var privateNetworks []*lb.PrivateNetwork

	for i := range pnConfigs {
		pn, err := lbAPI.AttachPrivateNetwork(&lb.ZonedAPIAttachPrivateNetworkRequest{
			Zone:             zone,
			LBID:             lbID,
			PrivateNetworkID: pnConfigs[i].PrivateNetworkID,
			StaticConfig:     pnConfigs[i].StaticConfig, //nolint:staticcheck
			DHCPConfig:       pnConfigs[i].DHCPConfig,   //nolint:staticcheck
			IpamIDs:          pnConfigs[i].IpamIDs,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return nil, err
		}

		privateNetworks, err = waitForPrivateNetworks(ctx, lbAPI, zone, pn.LB.ID, timeout)
		if err != nil && !httperrors.Is404(err) {
			return nil, err
		}

		for _, pn := range privateNetworks {
			if pn.Status == lb.PrivateNetworkStatusError {
				err = lbAPI.DetachPrivateNetwork(&lb.ZonedAPIDetachPrivateNetworkRequest{
					Zone:             zone,
					LBID:             pn.LB.ID,
					PrivateNetworkID: pn.PrivateNetworkID,
				}, scw.WithContext(ctx))
				if err != nil && !httperrors.Is404(err) {
					return nil, err
				}

				tflog.Debug(ctx, fmt.Sprintf("DHCP config: %v", pn.DHCPConfig))     //nolint:staticcheck
				tflog.Debug(ctx, fmt.Sprintf("Static config: %v", pn.StaticConfig)) //nolint:staticcheck

				return nil, fmt.Errorf("attaching private network with id: %s on error state. please check your config", pn.PrivateNetworkID)
			}
		}
	}

	return privateNetworks, nil
}

func flattenLbInstances(instances []*lb.Instance) any {
	if instances == nil {
		return nil
	}

	flattenedInstances := []map[string]any(nil)
	for _, instance := range instances {
		flattenedInstances = append(flattenedInstances, map[string]any{
			"id":         instance.ID,
			"status":     instance.Status.String(),
			"ip_address": instance.IPAddress,
			"created_at": types.FlattenTime(instance.CreatedAt),
			"updated_at": types.FlattenTime(instance.UpdatedAt),
			"zone":       instance.Zone,
		})
	}

	return flattenedInstances
}

func flattenLbIPs(ips []*lb.IP) any {
	if ips == nil {
		return nil
	}

	flattenedIPs := []map[string]any(nil)
	for _, ip := range ips {
		flattenedIPs = append(flattenedIPs, map[string]any{
			"id":              ip.ID,
			"ip_address":      ip.IPAddress,
			"reverse":         ip.Reverse,
			"organization_id": ip.OrganizationID,
			"project_id":      ip.ProjectID,
			"zone":            ip.Zone,
			"lb_id":           types.FlattenStringPtr(ip.LBID),
		})
	}

	return flattenedIPs
}

func flattenLBIPIDs(zone scw.Zone, ips []*lb.IP) []string {
	if ips == nil {
		return nil
	}

	flattenedIPs := make([]string, len(ips))
	for i, ip := range ips {
		flattenedIPs[i] = zonal.NewIDString(zone, ip.ID)
	}

	return flattenedIPs
}
