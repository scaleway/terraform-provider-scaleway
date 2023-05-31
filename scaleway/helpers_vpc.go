package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// vpcAPIWithZone returns a new VPC API and the zone for a Create request
func vpcAPIWithZone(d *schema.ResourceData, m interface{}) (*v1.API, scw.Zone, error) {
	meta := m.(*Meta)
	vpcAPI := v1.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcAPI, zone, err
}

// vpcAPIWithZoneAndID
func vpcAPIWithZoneAndID(m interface{}, id string) (*v1.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	vpcAPI := v1.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcAPI, zone, ID, err
}

// vpcAPIWithRegion returns a new VPC API and the region for a Create request
func vpcAPIWithRegion(d *schema.ResourceData, m interface{}) (*v2.API, scw.Region, error) {
	meta := m.(*Meta)
	vpcAPI := v2.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcAPI, region, err
}

// vpcAPIWithRegionAndID
func vpcAPIWithRegionAndID(m interface{}, id string) (*v2.API, scw.Region, string, error) {
	meta := m.(*Meta)
	vpcAPI := v2.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcAPI, region, ID, err
}

func vpcAPI(m interface{}) (*v1.API, error) {
	meta, ok := m.(*Meta)
	if !ok {
		return nil, fmt.Errorf("wrong type: %T", m)
	}

	return v1.NewAPI(meta.scwClient), nil
}

func expandSubnets(data interface{}) ([]scw.IPNet, error) {
	var ipNets []scw.IPNet
	for _, s := range data.([]interface{}) {
		if s == nil {
			s = ""
		}
		ipNet, err := expandIPNet(s.(string))
		if err != nil {
			return nil, err
		}
		ipNets = append(ipNets, ipNet)
	}

	return ipNets, nil
}

func flattenSubnets(subnets []scw.IPNet) *schema.Set {
	var rawSubnets []interface{}
	for _, s := range subnets {
		rawSubnets = append(rawSubnets, s.String())
	}
	return schema.NewSet(func(i interface{}) int {
		return StringHashcode(i.(string))
	}, rawSubnets)
}
