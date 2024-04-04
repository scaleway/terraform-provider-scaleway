package lb

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func flattenPrivateNetworkConfigs(privateNetworks []*lb.PrivateNetwork) interface{} {
	if len(privateNetworks) == 0 || privateNetworks == nil {
		return nil
	}

	pnI := []map[string]interface{}(nil)
	var dhcpConfigExist bool
	for _, pn := range privateNetworks {
		if pn.DHCPConfig != nil {
			dhcpConfigExist = true
		}
		pnRegion, err := pn.LB.Zone.Region()
		if err != nil {
			return diag.FromErr(err)
		}
		pnRegionalID := regional.NewIDString(pnRegion, pn.PrivateNetworkID)
		pnI = append(pnI, map[string]interface{}{
			"private_network_id": pnRegionalID,
			"dhcp_config":        dhcpConfigExist,
			"status":             pn.Status.String(),
			"zone":               pn.LB.Zone.String(),
			"static_config":      flattenLbPrivateNetworkStaticConfig(pn.StaticConfig),
		})
	}
	return pnI
}

func flattenLbACLMatch(match *lb.ACLMatch) interface{} {
	if match == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"ip_subnet":          types.FlattenSliceStringPtr(match.IPSubnet),
			"http_filter":        match.HTTPFilter.String(),
			"http_filter_value":  types.FlattenSliceStringPtr(match.HTTPFilterValue),
			"http_filter_option": match.HTTPFilterOption,
			"invert":             match.Invert,
		},
	}
}

func expandLbACLMatch(raw interface{}) *lb.ACLMatch {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})

	// scaleway api require ip subnet, so if we did not specify one, just put 0.0.0.0/0 instead
	ipSubnet := types.ExpandSliceStringPtr(rawMap["ip_subnet"].([]interface{}))
	if len(ipSubnet) == 0 {
		ipSubnet = []*string{types.ExpandStringPtr("0.0.0.0/0")}
	}

	return &lb.ACLMatch{
		IPSubnet:         ipSubnet,
		HTTPFilter:       lb.ACLHTTPFilter(rawMap["http_filter"].(string)),
		HTTPFilterValue:  types.ExpandSliceStringPtr(rawMap["http_filter_value"].([]interface{})),
		HTTPFilterOption: types.ExpandStringPtr(rawMap["http_filter_option"].(string)),
		Invert:           rawMap["invert"].(bool),
	}
}

func expandLbBackendMarkdownAction(raw interface{}) lb.OnMarkedDownAction {
	if raw == "none" {
		return lb.OnMarkedDownActionOnMarkedDownActionNone
	}
	return lb.OnMarkedDownAction(raw.(string))
}

func flattenLbProtocol(protocol lb.Protocol) interface{} {
	return protocol.String()
}

func expandLbProtocol(raw interface{}) lb.Protocol {
	return lb.Protocol(raw.(string))
}

func flattenLbForwardPortAlgorithm(algo lb.ForwardPortAlgorithm) interface{} {
	return algo.String()
}

func expandLbForwardPortAlgorithm(raw interface{}) lb.ForwardPortAlgorithm {
	return lb.ForwardPortAlgorithm(raw.(string))
}

func flattenLbStickySessionsType(t lb.StickySessionsType) interface{} {
	return t.String()
}

func expandLbStickySessionsType(raw interface{}) lb.StickySessionsType {
	return lb.StickySessionsType(raw.(string))
}

func flattenLbHCTCP(config *lb.HealthCheckTCPConfig) interface{} {
	if config == nil {
		return nil
	}
	return []map[string]interface{}{
		{},
	}
}

func expandLbHCTCP(raw interface{}) *lb.HealthCheckTCPConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	return &lb.HealthCheckTCPConfig{}
}

func flattenLbHCHTTP(config *lb.HealthCheckHTTPConfig) interface{} {
	if config == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"uri":         config.URI,
			"method":      config.Method,
			"code":        types.FlattenInt32Ptr(config.Code),
			"host_header": config.HostHeader,
		},
	}
}

func expandLbHCHTTP(raw interface{}) *lb.HealthCheckHTTPConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lb.HealthCheckHTTPConfig{
		URI:        rawMap["uri"].(string),
		Method:     rawMap["method"].(string),
		Code:       types.ExpandInt32Ptr(rawMap["code"]),
		HostHeader: rawMap["host_header"].(string),
	}
}

func flattenLbHCHTTPS(config *lb.HealthCheckHTTPSConfig) interface{} {
	if config == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"uri":         config.URI,
			"method":      config.Method,
			"code":        types.FlattenInt32Ptr(config.Code),
			"host_header": config.HostHeader,
			"sni":         config.Sni,
		},
	}
}

func expandLbHCHTTPS(raw interface{}) *lb.HealthCheckHTTPSConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lb.HealthCheckHTTPSConfig{
		URI:        rawMap["uri"].(string),
		Method:     rawMap["method"].(string),
		Code:       types.ExpandInt32Ptr(rawMap["code"]),
		HostHeader: rawMap["host_header"].(string),
		Sni:        rawMap["sni"].(string),
	}
}

func expandLbLetsEncrypt(raw interface{}) *lb.CreateCertificateRequestLetsencryptConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	alternativeNames := rawMap["subject_alternative_name"].([]interface{})
	config := &lb.CreateCertificateRequestLetsencryptConfig{
		CommonName: rawMap["common_name"].(string),
	}
	for _, alternativeName := range alternativeNames {
		config.SubjectAlternativeName = append(config.SubjectAlternativeName, alternativeName.(string))
	}
	return config
}

func expandLbCustomCertificate(raw interface{}) *lb.CreateCertificateRequestCustomCertificate {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	config := &lb.CreateCertificateRequestCustomCertificate{
		CertificateChain: rawMap["certificate_chain"].(string),
	}
	return config
}

func expandLbProxyProtocol(raw interface{}) lb.ProxyProtocol {
	return lb.ProxyProtocol("proxy_protocol_" + raw.(string))
}

func flattenLbProxyProtocol(pp lb.ProxyProtocol) interface{} {
	return strings.TrimPrefix(pp.String(), "proxy_protocol_")
}

func flattenLbBackendMarkdownAction(action lb.OnMarkedDownAction) interface{} {
	if action == lb.OnMarkedDownActionOnMarkedDownActionNone {
		return "none"
	}
	return action.String()
}

func flattenLbACL(acl *lb.ACL) interface{} {
	res := map[string]interface{}{
		"name":   acl.Name,
		"match":  flattenLbACLMatch(acl.Match),
		"action": flattenLbACLAction(acl.Action),
	}
	return res
}

// expandLbACL transforms a state acl to an api one.
func expandLbACL(i interface{}) *lb.ACL {
	rawRule := i.(map[string]interface{})
	acl := &lb.ACL{
		Name:        rawRule["name"].(string),
		Description: rawRule["description"].(string),
		Match:       expandLbACLMatch(rawRule["match"]),
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

func flattenLbACLAction(action *lb.ACLAction) interface{} {
	return []map[string]interface{}{
		{
			"type":     action.Type,
			"redirect": flattenLbACLActionRedirect(action.Redirect),
		},
	}
}

func expandLbACLAction(raw interface{}) *lb.ACLAction {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lb.ACLAction{
		Type:     lb.ACLActionType(rawMap["type"].(string)),
		Redirect: expandLbACLActionRedirect(rawMap["redirect"]),
	}
}

func flattenLbACLActionRedirect(redirect *lb.ACLActionRedirect) interface{} {
	if redirect == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"type":   redirect.Type,
			"target": redirect.Target,
			"code":   redirect.Code,
		},
	}
}

func expandLbACLActionRedirect(raw interface{}) *lb.ACLActionRedirect {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lb.ACLActionRedirect{
		Type:   lb.ACLActionRedirectRedirectType(rawMap["type"].(string)),
		Target: rawMap["target"].(string),
		Code:   types.ExpandInt32Ptr(rawMap["code"]),
	}
}

func expandPrivateNetworks(data interface{}) ([]*lb.PrivateNetwork, error) {
	if data == nil {
		return nil, nil
	}

	pns := []*lb.PrivateNetwork(nil)

	for _, pn := range data.(*schema.Set).List() {
		rawPn := pn.(map[string]interface{})
		privateNetwork := &lb.PrivateNetwork{}
		privateNetwork.PrivateNetworkID = locality.ExpandID(rawPn["private_network_id"].(string))
		if staticConfig, hasStaticConfig := rawPn["static_config"]; hasStaticConfig && len(staticConfig.([]interface{})) > 0 {
			privateNetwork.StaticConfig = expandLbPrivateNetworkStaticConfig(staticConfig)
		} else {
			privateNetwork.DHCPConfig = expandLbPrivateNetworkDHCPConfig(rawPn["dhcp_config"])
		}

		pns = append(pns, privateNetwork)
	}
	return pns, nil
}

func expandLbPrivateNetworkStaticConfig(raw interface{}) *lb.PrivateNetworkStaticConfig {
	if raw == nil || len(raw.([]interface{})) < 1 {
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

	return *cfg.IPAddress
}

func expandLbPrivateNetworkDHCPConfig(raw interface{}) *lb.PrivateNetworkDHCPConfig {
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
			StaticConfig:     pnConfigs[i].StaticConfig,
			DHCPConfig:       pnConfigs[i].DHCPConfig,
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
				tflog.Debug(ctx, fmt.Sprintf("DHCP config: %v", pn.DHCPConfig))
				tflog.Debug(ctx, fmt.Sprintf("Static config: %v", pn.StaticConfig))
				return nil, fmt.Errorf("attaching private network with id: %s on error state. please check your config", pn.PrivateNetworkID)
			}
		}
	}

	return privateNetworks, nil
}

func flattenLbInstances(instances []*lb.Instance) interface{} {
	if instances == nil {
		return nil
	}
	flattenedInstances := []map[string]interface{}(nil)
	for _, instance := range instances {
		flattenedInstances = append(flattenedInstances, map[string]interface{}{
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

func flattenLbIPs(ips []*lb.IP) interface{} {
	if ips == nil {
		return nil
	}
	flattenedIPs := []map[string]interface{}(nil)
	for _, ip := range ips {
		flattenedIPs = append(flattenedIPs, map[string]interface{}{
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
