package redis

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultRedisClusterTimeout           = 15 * time.Minute
	defaultWaitRedisClusterRetryInterval = 5 * time.Second
)

// newRedisApi returns a new Redis API
func newAPI(m any) *redis.API {
	return redis.NewAPI(meta.ExtractScwClient(m))
}

// newAPIWithZone returns a new Redis API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m any) (*redis.API, scw.Zone, error) {
	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return newAPI(m), zone, nil
}

// NewAPIWithZoneAndID returns a Redis API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, id string) (*redis.API, scw.Zone, string, error) {
	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return newAPI(m), zone, ID, nil
}

func waitForCluster(ctx context.Context, api *redis.API, zone scw.Zone, id string, timeout time.Duration) (*redis.Cluster, error) {
	retryInterval := defaultWaitRedisClusterRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForCluster(&redis.WaitForClusterRequest{
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		ClusterID:     id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func privateNetworkSetHash(v any) int {
	var buf bytes.Buffer

	m := v.(map[string]any)
	if pnID, ok := m["id"]; ok {
		buf.WriteString(locality.ExpandID(pnID))
	}

	if serviceIPs, ok := m["service_ips"]; ok {
		// Sort the service IPs before generating the hash.
		ips := serviceIPs.([]any)
		sort.Slice(ips, func(i, j int) bool {
			return ips[i].(string) < ips[j].(string)
		})

		for i, item := range ips {
			buf.WriteString(fmt.Sprintf("%d-%s-", i, item.(string)))
		}
	}

	return types.StringHashcode(buf.String())
}
