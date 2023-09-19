package scaleway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	telemetryReporting                 = "telemetry_reporting"
	defaultDocumentDBInstanceTimeout   = defaultRdbInstanceTimeout
	defaultWaitDocumentDBRetryInterval = defaultWaitRDBRetryInterval
)

// documentDBAPIWithRegion returns a new documentdb API and the region for a Create request
func documentDBAPIWithRegion(d *schema.ResourceData, m interface{}) (*documentdb.API, scw.Region, error) {
	meta := m.(*Meta)
	api := documentdb.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// documentDBAPIWithRegionalAndID returns a new documentdb API with region and ID extracted from the state
func documentDBAPIWithRegionAndID(m interface{}, regionalID string) (*documentdb.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := documentdb.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func waitForDocumentDBInstance(ctx context.Context, api *documentdb.API, region scw.Region, id string, timeout time.Duration) (*documentdb.Instance, error) {
	retryInterval := defaultWaitDocumentDBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	instance, err := api.WaitForInstance(&documentdb.WaitForInstanceRequest{
		Region:        region,
		InstanceID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return instance, err
}

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/DatabaseName"
func resourceScalewayDocumentDBDatabaseID(region scw.Region, instanceID string, databaseName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, databaseName)
}

// resourceScalewayDocumentDBDatabaseName extract regional instanceID and databaseName from composed ID
// returned by resourceScalewayDocumentDBDatabaseID()
func resourceScalewayDocumentDBDatabaseName(id string) (string, string, error) {
	elems := strings.Split(id, "/")
	if len(elems) != 3 {
		return "", "", fmt.Errorf("cant parse terraform database id: %s", id)
	}

	return elems[0] + "/" + elems[1], elems[2], nil
}

func expandDocumentDBPrivateNetwork(data interface{}, exist bool) (*documentdb.EndpointSpecPrivateNetwork, error) {
	if data == nil || !exist {
		return nil, nil
	}

	var res *documentdb.EndpointSpecPrivateNetwork
	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		ipNet := r["ip_net"].(string)
		res = &documentdb.EndpointSpecPrivateNetwork{
			PrivateNetworkID: expandID(r["id"].(string)),
		}
		if len(ipNet) > 0 {
			ip, err := expandIPNet(r["ip_net"].(string))
			if err != nil {
				return res, err
			}
			res.ServiceIP = &ip
		} else {
			res.IpamConfig = &documentdb.EndpointSpecPrivateNetworkIpamConfig{}
		}
	}

	return res, nil
}

func flattenDocumentDBPrivateNetwork(pn *documentdb.EndpointPrivateNetworkDetails) (interface{}, error) {
	pnI := []map[string]interface{}(nil)
	fetchRegion, err := pn.Zone.Region()
	if err != nil {
		return nil, err
	}

	pnID := newRegionalIDString(fetchRegion, pn.PrivateNetworkID)
	serviceIP, err := flattenIPNet(pn.ServiceIP)
	if err != nil {
		return pnI, err
	}
	pnI = append(pnI, map[string]interface{}{
		"ip_net": serviceIP,
		"id":     pnID,
		"zone":   pn.Zone.String(),
	})

	return pnI, nil
}

func flattenDocumentDBLoadBalancer(endpoints []*documentdb.Endpoint) interface{} {
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
