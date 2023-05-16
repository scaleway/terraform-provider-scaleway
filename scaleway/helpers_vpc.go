package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// vpcAPIWithZone returns a new VPC API and the zone for a Create request
func vpcAPIWithZone(d *schema.ResourceData, m interface{}) (*vpc.API, scw.Zone, error) {
	meta := m.(*Meta)
	vpcAPI := vpc.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcAPI, zone, err
}

// vpcAPIWithZoneAndID
func vpcAPIWithZoneAndID(m interface{}, id string) (*vpc.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	vpcAPI := vpc.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcAPI, zone, ID, err
}

func vpcAPI(m interface{}) (*vpc.API, error) {
	meta, ok := m.(*Meta)
	if !ok {
		return nil, fmt.Errorf("wrong type: %T", m)
	}

	return vpc.NewAPI(meta.scwClient), nil
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
