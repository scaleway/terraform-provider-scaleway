package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// getLbAPI returns a new lb API
func getLbAPI(m interface{}) *lb.API {
	meta := m.(*Meta)
	return lb.NewAPI(meta.scwClient)
}

// getLbAPIWithRegion returns a new lb API and the region for a Create request
func getLbAPIWithRegion(d *schema.ResourceData, m interface{}) (*lb.API, scw.Region, error) {
	meta := m.(*Meta)
	lbApi := lb.NewAPI(meta.scwClient)

	region, err := getRegion(d, meta)
	return lbApi, region, err
}

// getLbAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func getLbAPIWithRegionAndID(m interface{}, id string) (*lb.API, scw.Region, string, error) {
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
	if t == lb.StickySessionsTypeNone {
		return "none"
	}
	return t.String()
}

func expandLbStickySessionsType(raw interface{}) lb.StickySessionsType {
	if raw == "none" {
		return lb.StickySessionsTypeNone
	}
	return lb.StickySessionsType(raw.(string))
}
