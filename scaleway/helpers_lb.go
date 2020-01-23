package scaleway

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	LbWaitForTimeout = 10 * time.Minute
)

// lbAPI returns a new lb API
func lbAPI(m interface{}) *lb.API {
	meta := m.(*Meta)
	return lb.NewAPI(meta.scwClient)
}

// lbAPIWithRegion returns a new lb API and the region for a Create request
func lbAPIWithRegion(d *schema.ResourceData, m interface{}) (*lb.API, scw.Region, error) {
	meta := m.(*Meta)
	lbApi := lb.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	return lbApi, region, err
}

// lbAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func lbAPIWithRegionAndID(m interface{}, id string) (*lb.API, scw.Region, string, error) {
	meta := m.(*Meta)
	lbApi := lb.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	return lbApi, region, ID, err
}

func flattenLbBackendMarkdownAction(action lb.OnMarkedDownAction) interface{} {
	if action == lb.OnMarkedDownActionOnMarkedDownActionNone {
		return "none"
	}
	return action.String()
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
