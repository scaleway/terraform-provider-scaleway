package scaleway

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	validator "github.com/scaleway/scaleway-sdk-go/validation"
)

const (
	defaultLbLbTimeout = 10 * time.Minute
	retryLbIPInterval  = 5 * time.Second
)

// lbAPIWithZone returns an lb API WITH zone for a Create request
func lbAPIWithZone(d *schema.ResourceData, m interface{}) (*lbSDK.ZonedAPI, scw.Zone, error) {
	meta := m.(*Meta)
	lbAPI := lbSDK.NewZonedAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return lbAPI, zone, nil
}

// lbAPIWithZoneAndID returns an lb API with zone and ID extracted from the state
func lbAPIWithZoneAndID(m interface{}, id string) (*lbSDK.ZonedAPI, scw.Zone, string, error) {
	meta := m.(*Meta)
	lbAPI := lbSDK.NewZonedAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return lbAPI, zone, ID, nil
}

func flattenLbBackendMarkdownAction(action lbSDK.OnMarkedDownAction) interface{} {
	if action == lbSDK.OnMarkedDownActionOnMarkedDownActionNone {
		return "none"
	}
	return action.String()
}

func flattenLbACL(acl *lbSDK.ACL) interface{} {
	res := map[string]interface{}{
		"name":   acl.Name,
		"match":  flattenLbACLMatch(acl.Match),
		"action": flattenLbACLAction(acl.Action),
	}
	return res
}

// expandLbACL transforms a state acl to an api one.
func expandLbACL(i interface{}) *lbSDK.ACL {
	rawRule := i.(map[string]interface{})
	acl := &lbSDK.ACL{
		Name:   rawRule["name"].(string),
		Match:  expandLbACLMatch(rawRule["match"]),
		Action: expandLbACLAction(rawRule["action"]),
	}

	//remove http filter values if we do not pass any http filter
	if acl.Match.HTTPFilter == "" || acl.Match.HTTPFilter == lbSDK.ACLHTTPFilterACLHTTPFilterNone {
		acl.Match.HTTPFilter = lbSDK.ACLHTTPFilterACLHTTPFilterNone
		acl.Match.HTTPFilterValue = []*string{}
	}

	return acl
}
func flattenLbACLAction(action *lbSDK.ACLAction) interface{} {
	return []map[string]interface{}{
		{
			"type": action.Type,
		},
	}
}

func expandPrivateNetworks(data interface{}, lbID string) ([]*lbSDK.ZonedAPIAttachPrivateNetworkRequest, error) {
	if data == nil {
		return nil, nil
	}

	var res []*lbSDK.ZonedAPIAttachPrivateNetworkRequest
	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		zonePN, pnID, err := parseZonedID(r["private_network_id"].(string))
		if err != nil {
			return nil, err
		}
		pnRequest := &lbSDK.ZonedAPIAttachPrivateNetworkRequest{
			PrivateNetworkID: pnID,
			Zone:             zonePN,
			LBID:             lbID,
		}

		staticConfig := r["static_config"]
		if len(staticConfig.([]interface{})) > 0 {
			pnRequest.StaticConfig = expandLbPrivateNetworkStaticConfig(staticConfig)
		} else {
			pnRequest.DHCPConfig = expandLbPrivateNetworkDHCPConfig(r["dhcp_config"])
		}

		res = append(res, pnRequest)
	}

	return res, nil
}

func isPrivateNetworkEqual(A, B interface{}) bool {
	// Find out the diff Private Network or not
	if _, ok := A.(*lbSDK.PrivateNetwork); ok {
		if _, ok := B.(*lbSDK.PrivateNetwork); ok {
			if A.(*lbSDK.PrivateNetwork).PrivateNetworkID == B.(*lbSDK.PrivateNetwork).PrivateNetworkID {
				// if both has dhcp config should not update
				if A.(*lbSDK.PrivateNetwork).DHCPConfig != nil && B.(*lbSDK.PrivateNetwork).DHCPConfig != nil {
					return true
				}
				// check static config
				aConfig := A.(*lbSDK.PrivateNetwork).StaticConfig
				bConfig := B.(*lbSDK.PrivateNetwork).StaticConfig
				if aConfig != nil && bConfig != nil {
					// check if static config is different
					return reflect.DeepEqual(aConfig.IPAddress, bConfig.IPAddress)
				}
			}
		}
	}
	return false
}

func newPrivateNetwork(raw map[string]interface{}) *lbSDK.PrivateNetwork {
	_, pnID, _ := parseZonedID(raw["private_network_id"].(string))

	pn := &lbSDK.PrivateNetwork{PrivateNetworkID: pnID}
	staticConfig := raw["static_config"]
	if len(staticConfig.([]interface{})) > 0 {
		pn.StaticConfig = expandLbPrivateNetworkStaticConfig(staticConfig)
	} else {
		pn.DHCPConfig = expandLbPrivateNetworkDHCPConfig(raw["dhcp_config"])
	}

	return pn
}
func privateNetworksToDetach(pns []*lbSDK.PrivateNetwork, updates interface{}) (map[string]bool, error) {
	actions := make(map[string]bool, len(pns))
	configs := make(map[string]*lbSDK.PrivateNetwork, len(pns))
	// set detached all as default
	for _, pn := range pns {
		actions[pn.PrivateNetworkID] = true
		configs[pn.PrivateNetworkID] = pn
	}
	//check if private network still exist or is different
	for _, pn := range updates.([]interface{}) {
		r := pn.(map[string]interface{})
		_, pnID, err := parseZonedID(r["private_network_id"].(string))
		if err != nil {
			return nil, err
		}
		if _, exist := actions[pnID]; exist {
			// check if config are equal
			actions[pnID] = !isPrivateNetworkEqual(configs[pnID], newPrivateNetwork(r))
		}
	}
	return actions, nil
}

func flattenPrivateNetworkConfigs(resList *lbSDK.ListLBPrivateNetworksResponse) interface{} {
	if len(resList.PrivateNetwork) == 0 || resList == nil {
		return nil
	}

	pnConfigs := resList.PrivateNetwork
	pnI := []map[string]interface{}(nil)
	var dhcpConfigExist bool
	for _, pn := range pnConfigs {
		if pn.DHCPConfig != nil {
			dhcpConfigExist = true
		}
		pnZonedID := newZonedIDString(pn.LB.Zone, pn.PrivateNetworkID)
		pnI = append(pnI, map[string]interface{}{
			"private_network_id": pnZonedID,
			"dhcp_config":        dhcpConfigExist,
			"status":             pn.Status.String(),
			"zone":               pn.LB.Zone.String(),
			"static_config":      flattenLbPrivateNetworkStaticConfig(pn.StaticConfig),
		})
	}

	return pnI
}

func expandLbACLAction(raw interface{}) *lbSDK.ACLAction {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lbSDK.ACLAction{
		Type: lbSDK.ACLActionType(rawMap["type"].(string)),
	}
}

func flattenLbACLMatch(match *lbSDK.ACLMatch) interface{} {
	return []map[string]interface{}{
		{
			"ip_subnet":          flattenSliceStringPtr(match.IPSubnet),
			"http_filter":        match.HTTPFilter.String(),
			"http_filter_value":  flattenSliceStringPtr(match.HTTPFilterValue),
			"http_filter_option": match.HTTPFilterOption,
			"invert":             match.Invert,
		},
	}
}

func expandLbACLMatch(raw interface{}) *lbSDK.ACLMatch {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})

	//scaleway api require ip subnet, so if we did not specify one, just put 0.0.0.0/0 instead
	ipSubnet := expandSliceStringPtr(rawMap["ip_subnet"].([]interface{}))
	if len(ipSubnet) == 0 {
		ipSubnet = []*string{expandStringPtr("0.0.0.0/0")}
	}

	return &lbSDK.ACLMatch{
		IPSubnet:         ipSubnet,
		HTTPFilter:       lbSDK.ACLHTTPFilter(rawMap["http_filter"].(string)),
		HTTPFilterValue:  expandSliceStringPtr(rawMap["http_filter_value"].([]interface{})),
		HTTPFilterOption: expandStringPtr(rawMap["http_filter_option"].(string)),
		Invert:           rawMap["invert"].(bool),
	}
}

func expandLbBackendMarkdownAction(raw interface{}) lbSDK.OnMarkedDownAction {
	if raw == "none" {
		return lbSDK.OnMarkedDownActionOnMarkedDownActionNone
	}
	return lbSDK.OnMarkedDownAction(raw.(string))
}

func flattenLbProtocol(protocol lbSDK.Protocol) interface{} {
	return protocol.String()
}

func expandLbProtocol(raw interface{}) lbSDK.Protocol {
	return lbSDK.Protocol(raw.(string))
}

func flattenLbForwardPortAlgorithm(algo lbSDK.ForwardPortAlgorithm) interface{} {
	return algo.String()
}

func expandLbForwardPortAlgorithm(raw interface{}) lbSDK.ForwardPortAlgorithm {
	return lbSDK.ForwardPortAlgorithm(raw.(string))
}

func flattenLbStickySessionsType(t lbSDK.StickySessionsType) interface{} {
	return t.String()
}

func expandLbStickySessionsType(raw interface{}) lbSDK.StickySessionsType {
	return lbSDK.StickySessionsType(raw.(string))
}

func flattenLbHCTCP(config *lbSDK.HealthCheckTCPConfig) interface{} {
	if config == nil {
		return nil
	}
	return []map[string]interface{}{
		{},
	}
}

func expandLbHCTCP(raw interface{}) *lbSDK.HealthCheckTCPConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	return &lbSDK.HealthCheckTCPConfig{}
}

func flattenLbHCHTTP(config *lbSDK.HealthCheckHTTPConfig) interface{} {
	if config == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"uri":    config.URI,
			"method": config.Method,
			"code":   flattenInt32Ptr(config.Code),
		},
	}
}

func expandLbHCHTTP(raw interface{}) *lbSDK.HealthCheckHTTPConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lbSDK.HealthCheckHTTPConfig{
		URI:    rawMap["uri"].(string),
		Method: rawMap["method"].(string),
		Code:   expandInt32Ptr(rawMap["code"]),
	}
}

func flattenLbHCHTTPS(config *lbSDK.HealthCheckHTTPSConfig) interface{} {
	if config == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"uri":    config.URI,
			"method": config.Method,
			"code":   flattenInt32Ptr(config.Code),
		},
	}
}

func expandLbHCHTTPS(raw interface{}) *lbSDK.HealthCheckHTTPSConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lbSDK.HealthCheckHTTPSConfig{
		URI:    rawMap["uri"].(string),
		Method: rawMap["method"].(string),
		Code:   expandInt32Ptr(rawMap["code"]),
	}
}

func expandLbLetsEncrypt(raw interface{}) *lbSDK.CreateCertificateRequestLetsencryptConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	alternativeNames := rawMap["subject_alternative_name"].([]interface{})
	config := &lbSDK.CreateCertificateRequestLetsencryptConfig{
		CommonName: rawMap["common_name"].(string),
	}
	for _, alternativeName := range alternativeNames {
		config.SubjectAlternativeName = append(config.SubjectAlternativeName, alternativeName.(string))
	}
	return config
}

func expandLbCustomCertificate(raw interface{}) *lbSDK.CreateCertificateRequestCustomCertificate {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	config := &lbSDK.CreateCertificateRequestCustomCertificate{
		CertificateChain: rawMap["certificate_chain"].(string),
	}
	return config
}

func expandLbProxyProtocol(raw interface{}) lbSDK.ProxyProtocol {
	return lbSDK.ProxyProtocol("proxy_protocol_" + raw.(string))
}

func flattenLbProxyProtocol(pp lbSDK.ProxyProtocol) interface{} {
	return strings.TrimPrefix(pp.String(), "proxy_protocol_")
}

func lbUpgradeV1SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"id": cty.String,
	})
}

// lbUpgradeV1UpgradeFunc allow upgrade the from regional to a zoned resource.
func lbUpgradeV1SchemaUpgradeFunc(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	var err error
	// element id: upgrade
	ID, exist := rawState["id"]
	if !exist {
		return nil, fmt.Errorf("upgrade: id not exist")
	}
	rawState["id"], err = lbUpgradeV1RegionalToZonedID(ID.(string))
	if err != nil {
		return nil, err
	}
	// return rawState updated
	return rawState, nil
}

func lbUpgradeV1RegionalToZonedID(element string) (string, error) {
	locality, id, err := parseLocalizedID(element)
	// return error if can't parse
	if err != nil {
		return "", fmt.Errorf("upgrade: could not retrieve the locality from `%s`", element)
	}
	// if locality is already zoned return
	if validator.IsZone(locality) {
		return element, nil
	}
	//  append zone 1 as default: e.g. fr-par-1
	return fmt.Sprintf("%s-1/%s", locality, id), nil
}

func expandLbPrivateNetworkStaticConfig(raw interface{}) *lbSDK.PrivateNetworkStaticConfig {
	if raw == nil || len(raw.([]interface{})) < 1 {
		return nil
	}
	return &lbSDK.PrivateNetworkStaticConfig{
		IPAddress: expandStrings(raw),
	}
}

func flattenLbPrivateNetworkStaticConfig(cfg *lbSDK.PrivateNetworkStaticConfig) []string {
	if cfg == nil {
		return nil
	}

	return cfg.IPAddress
}

func expandLbPrivateNetworkDHCPConfig(raw interface{}) *lbSDK.PrivateNetworkDHCPConfig {
	if raw == nil || !raw.(bool) {
		return nil
	}
	return &lbSDK.PrivateNetworkDHCPConfig{}
}

func waitForLB(ctx context.Context, lbAPI *lbSDK.ZonedAPI, zone scw.Zone, LbID string, timeout time.Duration) (*lbSDK.LB, error) {
	retryInterval := defaultWaitLBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	loadBalancer, err := lbAPI.WaitForLb(&lbSDK.ZonedAPIWaitForLBRequest{
		LBID:          LbID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

func waitForLbInstances(ctx context.Context, lbAPI *lbSDK.ZonedAPI, zone scw.Zone, lbID string, timeout time.Duration) (*lbSDK.LB, error) {
	retryInterval := defaultWaitLBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	loadBalancer, err := lbAPI.WaitForLbInstances(&lbSDK.ZonedAPIWaitForLBInstancesRequest{
		Zone:          zone,
		LBID:          lbID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

func waitForLBPN(ctx context.Context, lbAPI *lbSDK.ZonedAPI, zone scw.Zone, LbID string, timeout time.Duration) ([]*lbSDK.PrivateNetwork, error) {
	retryInterval := defaultWaitLBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	privateNetworks, err := lbAPI.WaitForLBPN(&lbSDK.ZonedAPIWaitForLBPNRequest{
		LBID:          LbID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return privateNetworks, err
}

func waitForLBCertificate(ctx context.Context, lbAPI *lbSDK.ZonedAPI, zone scw.Zone, id string, timeout time.Duration) (*lbSDK.Certificate, error) {
	retryInterval := defaultWaitLBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	certificate, err := lbAPI.WaitForLBCertificate(&lbSDK.ZonedAPIWaitForLBCertificateRequest{
		CertID:        id,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return certificate, err
}
