package scaleway

import (
	"context"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultRdbInstanceTimeout = 45 * time.Minute
	createDatabaseTimeout     = 2 * time.Minute
)

// newRdbAPI returns a new RDB API
func newRdbAPI(m interface{}) *rdb.API {
	meta := m.(*Meta)
	return rdb.NewAPI(meta.scwClient)
}

// rdbAPIWithRegion returns a new lb API and the region for a Create request
func rdbAPIWithRegion(d *schema.ResourceData, m interface{}) (*rdb.API, scw.Region, error) {
	meta := m.(*Meta)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return newRdbAPI(m), region, nil
}

// rdbAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func rdbAPIWithRegionAndID(m interface{}, id string) (*rdb.API, scw.Region, string, error) {
	region, ID, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return newRdbAPI(m), region, ID, nil
}

func flattenRdbInstanceReadReplicas(readReplicas []*rdb.Endpoint) interface{} {
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

func flattenInstanceSettings(settings []*rdb.InstanceSetting) interface{} {
	res := make(map[string]string)
	for _, value := range settings {
		res[value.Name] = value.Value
	}

	return res
}

func expandInstanceSettings(i interface{}) []*rdb.InstanceSetting {
	rawRule := i.(map[string]interface{})
	var res []*rdb.InstanceSetting
	for key, value := range rawRule {
		res = append(res, &rdb.InstanceSetting{
			Name:  key,
			Value: value.(string),
		})
	}

	return res
}

func waitInstance(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.Instance, error) {
	retryInterval := defaultWaitRDBRetryInterval
	return api.WaitForInstance(&rdb.WaitForInstanceRequest{
		Region:        region,
		InstanceID:    id,
		Timeout:       scw.TimeDurationPtr(timeout), // upgrade takes some time
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func expandPrivateNetwork(data interface{}, exist bool) []*rdb.EndpointSpec {
	if data == nil || !exist {
		return nil
	}

	var res []*rdb.EndpointSpec
	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		spec := &rdb.EndpointSpec{
			PrivateNetwork: &rdb.EndpointSpecPrivateNetwork{
				PrivateNetworkID: expandID(r["pn_id"].(string)),
				ServiceIP:        expandIPNet(r["ip_net"].(string)),
			},
		}
		res = append(res, spec)
	}

	return res
}

func expandLoadBalancer() []*rdb.EndpointSpec {
	var res []*rdb.EndpointSpec

	res = append(res, &rdb.EndpointSpec{
		LoadBalancer: &rdb.EndpointSpecLoadBalancer{}})

	return res
}

func endpointsToRemove(endPoints []*rdb.Endpoint, updates interface{}) (map[string]bool, error) {
	actions := make(map[string]bool)
	endpoints := make(map[string]*rdb.Endpoint)
	for _, e := range endPoints {
		// skip load balancer
		if e.PrivateNetwork != nil {
			actions[e.ID] = true
			endpoints[newZonedIDString(e.PrivateNetwork.Zone, e.PrivateNetwork.PrivateNetworkID)] = e
		}
	}

	// compare if private networks are persisted
	for _, raw := range updates.([]interface{}) {
		r := raw.(map[string]interface{})
		pnZonedID := r["pn_id"].(string)
		locality, id, err := parseLocalizedID(pnZonedID)
		if err != nil {
			return nil, err
		}

		pnUpdated := newEndPointPrivateNetworkDetails(id, r["ip_net"].(string), locality)
		endpoint, exist := endpoints[pnZonedID]
		if !exist {
			continue
		}
		// match the endpoint id for a private network
		actions[endpoint.ID] = !isEndPointEqual(endpoints[pnZonedID].PrivateNetwork, pnUpdated)
	}

	return actions, nil
}

func newEndPointPrivateNetworkDetails(id, ip, locality string) *rdb.EndpointPrivateNetworkDetails {
	return &rdb.EndpointPrivateNetworkDetails{
		PrivateNetworkID: id,
		ServiceIP:        expandIPNet(ip),
		Zone:             scw.Zone(locality),
	}
}

func isEndPointEqual(A, B interface{}) bool {
	// Find out the diff Private Network or not
	if _, ok := A.(*rdb.EndpointPrivateNetworkDetails); ok {
		if _, ok := B.(*rdb.EndpointPrivateNetworkDetails); ok {
			detailsA := A.(*rdb.EndpointPrivateNetworkDetails)
			detailsB := B.(*rdb.EndpointPrivateNetworkDetails)
			return reflect.DeepEqual(detailsA, detailsB)
		}
	}
	return false
}

func flattenPrivateNetwork(endpoints []*rdb.Endpoint) (interface{}, bool) {
	pnI := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork != nil {
			pn := endpoint.PrivateNetwork
			pnZonedID := newZonedIDString(pn.Zone, pn.PrivateNetworkID)
			pnI = append(pnI, map[string]interface{}{
				"endpoint_id": endpoint.ID,
				"ip":          flattenIPPtr(endpoint.IP),
				"port":        int(endpoint.Port),
				"name":        endpoint.Name,
				"ip_net":      flattenIPNet(pn.ServiceIP),
				"pn_id":       pnZonedID,
				"hostname":    flattenStringPtr(endpoint.Hostname),
			})
			return pnI, true
		}
	}

	return pnI, false
}

func flattenLoadBalancer(endpoints []*rdb.Endpoint) interface{} {
	flat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.LoadBalancer != nil {
			flat = append(flat, map[string]interface{}{
				"endpoint_id": endpoint.ID,
				"ip":          flattenIPPtr(endpoint.IP),
				"port":        int(endpoint.Port),
				"name":        endpoint.Name,
				"hostname":    flattenStringPtr(endpoint.Hostname),
			})
			return flat
		}
	}

	return flat
}
