package scaleway

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	validator "github.com/scaleway/scaleway-sdk-go/validation"
)

const (
	lbWaitForTimeout   = 10 * time.Minute
	defaultLbLbTimeout = 10 * time.Minute
	retryLbIPInterval  = 5 * time.Second
)

// lbAPIWithZone returns an lb API WITH zone for a Create request
func lbAPIWithZone(d *schema.ResourceData, m interface{}) (*lb.ZonedAPI, scw.Zone, error) {
	meta := m.(*Meta)
	lbAPI := lb.NewZonedAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return lbAPI, zone, nil
}

// lbAPIWithZoneAndID returns an lb API with zone and ID extracted from the state
func lbAPIWithZoneAndID(m interface{}, id string) (*lb.ZonedAPI, scw.Zone, string, error) {
	meta := m.(*Meta)
	lbAPI := lb.NewZonedAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return lbAPI, zone, ID, nil
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
		Name:   rawRule["name"].(string),
		Match:  expandLbACLMatch(rawRule["match"]),
		Action: expandLbACLAction(rawRule["action"]),
	}

	//remove http filter values if we do not pass any http filter
	if acl.Match.HTTPFilter == "" || acl.Match.HTTPFilter == lb.ACLHTTPFilterACLHTTPFilterNone {
		acl.Match.HTTPFilter = lb.ACLHTTPFilterACLHTTPFilterNone
		acl.Match.HTTPFilterValue = []*string{}
	}

	return acl
}
func flattenLbACLAction(action *lb.ACLAction) interface{} {
	return []map[string]interface{}{
		{
			"type": action.Type,
		},
	}
}

func expandPrivateNetworks(data interface{}, lbID string) ([]*lb.ZonedAPIAttachPrivateNetworkRequest, error) {
	if data == nil {
		return nil, nil
	}

	var res []*lb.ZonedAPIAttachPrivateNetworkRequest
	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		zonePN, pnID, err := parseZonedID(r["private_network_id"].(string))
		if err != nil {
			return nil, err
		}
		pnRequest := &lb.ZonedAPIAttachPrivateNetworkRequest{
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
	if _, ok := A.(*lb.PrivateNetwork); ok {
		if _, ok := B.(*lb.PrivateNetwork); ok {
			if A.(*lb.PrivateNetwork).PrivateNetworkID == B.(*lb.PrivateNetwork).PrivateNetworkID {
				// if both has dhcp config should not update
				if A.(*lb.PrivateNetwork).DHCPConfig != nil && B.(*lb.PrivateNetwork).DHCPConfig != nil {
					return true
				}
				// check static config
				aConfig := A.(*lb.PrivateNetwork).StaticConfig
				bConfig := B.(*lb.PrivateNetwork).StaticConfig
				if aConfig != nil && bConfig != nil {
					// check if static config is different
					return reflect.DeepEqual(aConfig.IPAddress, bConfig.IPAddress)
				}
			}
		}
	}
	return false
}

func newPrivateNetwork(raw map[string]interface{}) *lb.PrivateNetwork {
	_, pnID, _ := parseZonedID(raw["private_network_id"].(string))

	pn := &lb.PrivateNetwork{PrivateNetworkID: pnID}
	staticConfig := raw["static_config"]
	if len(staticConfig.([]interface{})) > 0 {
		pn.StaticConfig = expandLbPrivateNetworkStaticConfig(staticConfig)
	} else {
		pn.DHCPConfig = expandLbPrivateNetworkDHCPConfig(raw["dhcp_config"])
	}

	return pn
}
func privateNetworksToDetach(pns []*lb.PrivateNetwork, updates interface{}) (map[string]bool, error) {
	actions := make(map[string]bool, len(pns))
	configs := make(map[string]*lb.PrivateNetwork, len(pns))
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

func flattenPrivateNetworkConfigs(resList *lb.ListLBPrivateNetworksResponse) interface{} {
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
			"status":             pn.Status,
			"zone":               pn.LB.Zone,
			"static_config":      flattenLbPrivateNetworkStaticConfig(pn.StaticConfig),
		})
	}

	return pnI
}

func expandLbACLAction(raw interface{}) *lb.ACLAction {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lb.ACLAction{
		Type: lb.ACLActionType(rawMap["type"].(string)),
	}
}

func flattenLbACLMatch(match *lb.ACLMatch) interface{} {
	return []map[string]interface{}{
		{
			"ip_subnet":         flattenSliceStringPtr(match.IPSubnet),
			"http_filter":       match.HTTPFilter.String(),
			"http_filter_value": flattenSliceStringPtr(match.HTTPFilterValue),
			"invert":            match.Invert,
		},
	}
}

func expandLbACLMatch(raw interface{}) *lb.ACLMatch {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})

	//scaleway api require ip subnet, so if we did not specify one, just put 0.0.0.0/0 instead
	ipSubnet := expandSliceStringPtr(rawMap["ip_subnet"].([]interface{}))
	if len(ipSubnet) == 0 {
		ipSubnet = []*string{expandStringPtr("0.0.0.0/0")}
	}

	return &lb.ACLMatch{
		IPSubnet:        ipSubnet,
		HTTPFilter:      lb.ACLHTTPFilter(rawMap["http_filter"].(string)),
		HTTPFilterValue: expandSliceStringPtr(rawMap["http_filter_value"].([]interface{})),
		Invert:          rawMap["invert"].(bool),
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
			"uri":    config.URI,
			"method": config.Method,
			"code":   flattenInt32Ptr(config.Code),
		},
	}
}

func expandLbHCHTTP(raw interface{}) *lb.HealthCheckHTTPConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}
	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lb.HealthCheckHTTPConfig{
		URI:    rawMap["uri"].(string),
		Method: rawMap["method"].(string),
		Code:   expandInt32Ptr(rawMap["code"]),
	}
}

func flattenLbHCHTTPS(config *lb.HealthCheckHTTPSConfig) interface{} {
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

func expandLbHCHTTPS(raw interface{}) *lb.HealthCheckHTTPSConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &lb.HealthCheckHTTPSConfig{
		URI:    rawMap["uri"].(string),
		Method: rawMap["method"].(string),
		Code:   expandInt32Ptr(rawMap["code"]),
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

func expandLbPrivateNetworkStaticConfig(raw interface{}) *lb.PrivateNetworkStaticConfig {
	if raw == nil || len(raw.([]interface{})) < 1 {
		return nil
	}
	return &lb.PrivateNetworkStaticConfig{
		IPAddress: expandStrings(raw),
	}
}

func flattenLbPrivateNetworkStaticConfig(cfg *lb.PrivateNetworkStaticConfig) []string {
	if cfg == nil {
		return nil
	}

	return cfg.IPAddress
}

func expandLbPrivateNetworkDHCPConfig(raw interface{}) *lb.PrivateNetworkDHCPConfig {
	if raw == nil || !raw.(bool) {
		return nil
	}
	return &lb.PrivateNetworkDHCPConfig{}
}

func waitForLB(ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration) (*lb.LB, error) {
	lbAPI, _, err := lbAPIWithZone(d, meta)
	if err != nil {
		return nil, err
	}
	// parse lb_id. It will be forced to a zoned lb
	zone, LbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return nil, err
	}


	retryInterval := defaultWaitLBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	loadBalancer, err := lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID:          LbID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

<<<<<<< HEAD
func waitForLBCertificate(ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration) (*lb.Certificate, error) {
	lbAPI, _, err := lbAPIWithZone(d, meta)
	if err != nil {
		return nil, err
	}
	// parse lb_id. It will be forced to a zoned lb
	zone, LbID, err := parseZonedID(d.Get("lb_id").(string))
=======
func waitForLBBackend(ctx context.Context, d *schema.ResourceData, meta interface{}) (*lb.LB, error) {
	api, zone, _, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return nil, err
	}

	_, LbID, err := parseZonedID(d.Get("lb_id").(string))
>>>>>>> 46a6a6e7 (Refactor to enable easily the adding of timeout)
	if err != nil {
		return nil, err
	}

	retryInterval := defaultWaitLBRetryInterval

	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

<<<<<<< HEAD
	certificate, err := lbAPI.WaitForLBCertificate(&lb.ZonedAPIWaitForLBCertificateRequest{
		CertID:        LbID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
=======
	loadBalancer, err := api.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID:          LbID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(lbWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

func waitForLBFrontend(ctx context.Context, d *schema.ResourceData, meta interface{}) (*lb.LB, error) {
	api, zone, _, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return nil, err
	}

	_, LbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return nil, err
	}

	retryInterval := defaultWaitLBRetryInterval

	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	loadBalancer, err := api.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID:          LbID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(lbWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

func waitForLBCertificate(ctx context.Context, d *schema.ResourceData, meta interface{}) (*lb.Certificate, error) {
	api, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return nil, err
	}

	retryInterval := defaultWaitLBRetryInterval

	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	certificate, err := api.WaitForLBCertificate(&lb.ZonedAPIWaitForLBCertificateRequest{
		CertID:        ID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(defaultLbLbTimeout),
>>>>>>> 46a6a6e7 (Refactor to enable easily the adding of timeout)
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return certificate, err
}

<<<<<<< HEAD
func waitForLbInstances(ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration) (*lb.LB, error) {
	lbAPI, _, err := lbAPIWithZone(d, meta)
	if err != nil {
		return nil, err
	}
	// parse lb_id. It will be forced to a zoned lb
	zone, LbID, err := parseZonedID(d.Get("lb_id").(string))
=======
func waitForLbInstances(ctx context.Context, d *schema.ResourceData, meta interface{}) (*lb.LB, error) {
	api, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
>>>>>>> 46a6a6e7 (Refactor to enable easily the adding of timeout)
	if err != nil {
		return nil, err
	}

	retryInterval := defaultWaitLBRetryInterval

	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	loadBalancer, err := api.WaitForLbInstances(&lb.ZonedAPIWaitForLBInstancesRequest{
		Zone:          zone,
<<<<<<< HEAD
		LBID:          LbID,
		Timeout:       scw.TimeDurationPtr(timeout),
=======
		LBID:          ID,
		Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
>>>>>>> 46a6a6e7 (Refactor to enable easily the adding of timeout)
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

<<<<<<< HEAD
func waitForLBPN(ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration) ([]*lb.PrivateNetwork, error) {
	lbAPI, _, err := lbAPIWithZone(d, meta)
	if err != nil {
		return nil, err
	}
	// parse lb_id. It will be forced to a zoned lb
	zone, LbID, err := parseZonedID(d.Get("lb_id").(string))
=======
func waitForLBPN(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*lb.PrivateNetwork, error) {
	api, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
>>>>>>> 46a6a6e7 (Refactor to enable easily the adding of timeout)
	if err != nil {
		return nil, err
	}

	retryInterval := defaultWaitLBRetryInterval

	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	privateNetworks, err := api.WaitForLBPN(&lb.ZonedAPIWaitForLBPNRequest{
		Zone:          zone,
		LBID:          LbID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return privateNetworks, err
}
