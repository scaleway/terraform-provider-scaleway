package scaleway

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rdbV1 "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultRdbInstanceTimeout = 15 * time.Minute
)

// newRdbAPI returns a new RDB API
func newRdbAPI(m interface{}) *rdbV1.API {
	meta := m.(*Meta)
	return rdbV1.NewAPI(meta.scwClient)
}

// rdbAPIWithRegion returns a new lb API and the region for a Create request
func rdbAPIWithRegion(d *schema.ResourceData, m interface{}) (*rdbV1.API, scw.Region, error) {
	meta := m.(*Meta)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return newRdbAPI(m), region, nil
}

// rdbAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func rdbAPIWithRegionAndID(m interface{}, id string) (*rdbV1.API, scw.Region, string, error) {
	region, ID, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return newRdbAPI(m), region, ID, nil
}

func flattenRdbInstanceReadReplicas(readReplicas []*rdbV1.Endpoint) interface{} {
	replicasI := []map[string]interface{}(nil)
	for _, readReplica := range readReplicas {
		replicasI = append(replicasI, map[string]interface{}{
			"ip":   flattenIPPtr(readReplica.IP),
			"port": int(readReplica.Port),
			"name": flattenStringPtr(readReplica.Name),
		})
	}
	return replicasI
}

func flattenInstanceSettings(settings []*rdbV1.InstanceSetting) interface{} {
	res := make(map[string]string)
	for _, value := range settings {
		res[value.Name] = value.Value
	}

	return res
}

func expandInstanceSettings(i interface{}) []*rdbV1.InstanceSetting {
	rawRule := i.(map[string]interface{})
	var res []*rdbV1.InstanceSetting
	for key, value := range rawRule {
		res = append(res, &rdbV1.InstanceSetting{
			Name:  key,
			Value: value.(string),
		})
	}

	return res
}

func expandPrivateNetwork(data interface{}, exist bool) []*rdbV1.EndpointSpec {
	if data == nil || !exist {
		return nil
	}

	var res []*rdbV1.EndpointSpec
	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		spec := &rdbV1.EndpointSpec{
			PrivateNetwork: &rdbV1.EndpointSpecPrivateNetwork{
				PrivateNetworkID: expandID(r["pn_id"].(string)),
				ServiceIP:        expandIPNet(r["ip"].(string)),
			},
		}
		res = append(res, spec)
	}

	return res
}

func flattenInstancePrivateNetwork(readEndpoints []*rdbV1.Endpoint) interface{} {
	privateNetworkI := []map[string]interface{}(nil)
	if len(readEndpoints) == 0 {
		return nil
	}
	for _, readPN := range readEndpoints {
		if readPN.PrivateNetwork != nil {
			pn := readPN.PrivateNetwork
			privateNetworkI = append(privateNetworkI, map[string]interface{}{
				"ip":    flattenIPNet(pn.ServiceIP),
				"pn_id": pn.PrivateNetworkID,
			})
		}
	}
	return privateNetworkI
}
