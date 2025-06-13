package autoscaling

import (
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandInstanceCapacity(raw interface{}) *autoscaling.Capacity {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	capacity := &autoscaling.Capacity{
		MaxReplicas: uint32(rawMap["max_replicas"].(int)),
		MinReplicas: uint32(rawMap["min_replicas"].(int)),
	}

	if rawVal, ok := rawMap["cooldown_delay"].(int); ok {
		capacity.CooldownDelay = &scw.Duration{Seconds: int64(rawVal)}
	}

	return capacity
}

func expandUpdateInstanceCapacity(raw interface{}) *autoscaling.UpdateInstanceGroupRequestCapacity {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	capacity := &autoscaling.UpdateInstanceGroupRequestCapacity{}

	if rawVal, ok := rawMap["max_replicas"].(int); ok {
		capacity.MaxReplicas = types.ExpandUint32Ptr(rawVal)
	}

	if rawVal, ok := rawMap["min_replicas"].(int); ok {
		capacity.MinReplicas = types.ExpandUint32Ptr(rawVal)
	}

	if rawVal, ok := rawMap["cooldown_delay"].(int); ok {
		capacity.CooldownDelay = &scw.Duration{Seconds: int64(rawVal)}
	}

	return capacity
}

func expandInstanceLoadBalancer(raw interface{}) *autoscaling.Loadbalancer {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})

	return &autoscaling.Loadbalancer{
		ID:               locality.ExpandID(rawMap["id"].(string)),
		PrivateNetworkID: locality.ExpandID(rawMap["private_network_id"].(string)),
		BackendIDs:       locality.ExpandIDs(rawMap["backend_ids"]),
	}
}

func expandPolicyMetric(raw interface{}) *autoscaling.Metric {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})

	var managedPtr *autoscaling.MetricManagedMetric

	if s := rawMap["managed_metric"].(string); s != "" {
		m := autoscaling.MetricManagedMetric(s)
		managedPtr = &m
	}

	return &autoscaling.Metric{
		Name:              rawMap["name"].(string),
		ManagedMetric:     managedPtr,
		CockpitMetricName: types.ExpandStringPtr(rawMap["cockpit_metric_name"].(string)),
		Operator:          autoscaling.MetricOperator(rawMap["operator"].(string)),
		Aggregate:         autoscaling.MetricAggregate(rawMap["aggregate"].(string)),
		SamplingRangeMin:  uint32(rawMap["sampling_range_min"].(int)),
		Threshold:         float32(rawMap["threshold"].(int)),
	}
}

func expandUpdatePolicyMetric(raw interface{}) *autoscaling.UpdateInstancePolicyRequestMetric {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})

	var managedPtr *autoscaling.UpdateInstancePolicyRequestMetricManagedMetric

	if s := rawMap["managed_metric"].(string); s != "" {
		m := autoscaling.UpdateInstancePolicyRequestMetricManagedMetric(s)
		managedPtr = &m
	}

	return &autoscaling.UpdateInstancePolicyRequestMetric{
		Name:              types.ExpandStringPtr(rawMap["name"].(string)),
		ManagedMetric:     managedPtr,
		CockpitMetricName: types.ExpandStringPtr(rawMap["cockpit_metric_name"].(string)),
		Operator:          autoscaling.UpdateInstancePolicyRequestMetricOperator(rawMap["operator"].(string)),
		Aggregate:         autoscaling.UpdateInstancePolicyRequestMetricAggregate(rawMap["aggregate"].(string)),
		SamplingRangeMin:  types.ExpandUint32Ptr(rawMap["sampling_range_min"].(int)),
		Threshold:         scw.Float32Ptr(float32(rawMap["sampling_threshold"].(int))),
	}
}

func expandUpdateInstanceLoadBalancer(raw interface{}) *autoscaling.UpdateInstanceGroupRequestLoadbalancer {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	lb := &autoscaling.UpdateInstanceGroupRequestLoadbalancer{}

	if rawVal, ok := rawMap["backend_ids"].(int); ok {
		lb.BackendIDs = types.ExpandStringsPtr(locality.ExpandIDs(rawVal))
	}

	return lb
}

func expandVolumes(rawVols []interface{}) map[string]*autoscaling.VolumeInstanceTemplate {
	vols := make(map[string]*autoscaling.VolumeInstanceTemplate, len(rawVols))

	for i, raw := range rawVols {
		m := raw.(map[string]interface{})
		vt := &autoscaling.VolumeInstanceTemplate{
			Name:       m["name"].(string),
			Boot:       m["boot"].(bool),
			VolumeType: autoscaling.VolumeInstanceTemplateVolumeType(m["volume_type"].(string)),
			Tags:       types.ExpandStrings(m["tags"].([]interface{})),
			PerfIops:   types.ExpandUint32Ptr(m["perf_iops"].(int)),
		}

		if slice := m["from_empty"].([]interface{}); len(slice) > 0 && slice[0] != nil {
			inner := slice[0].(map[string]interface{})
			sizeGB := inner["size"].(int)
			vt.FromEmpty = &autoscaling.VolumeInstanceTemplateFromEmpty{
				Size: scw.Size(uint64(sizeGB) * uint64(scw.GB)),
			}
		}

		if slice := m["from_snapshot"].([]interface{}); len(slice) > 0 && slice[0] != nil {
			inner := slice[0].(map[string]interface{})
			snapshot := &autoscaling.VolumeInstanceTemplateFromSnapshot{
				SnapshotID: locality.ExpandID(inner["snapshot_id"].(string)),
			}

			if sz, ok := inner["size"].(int); ok && sz > 0 {
				snapshot.Size = scw.SizePtr(scw.Size(uint64(sz) * uint64(scw.GB)))
			}

			vt.FromSnapshot = snapshot
		}

		vols[strconv.Itoa(i)] = vt
	}

	return vols
}

func flattenInstanceCapacity(capacity *autoscaling.Capacity) interface{} {
	if capacity == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"max_replicas":   capacity.MaxReplicas,
			"min_replicas":   capacity.MinReplicas,
			"cooldown_delay": capacity.CooldownDelay.Seconds,
		},
	}
}

func flattenInstanceLoadBalancer(lb *autoscaling.Loadbalancer, zone scw.Zone) interface{} {
	if lb == nil {
		return nil
	}

	pnRegion, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	return []map[string]interface{}{
		{
			"id":                 zonal.NewIDString(zone, lb.ID),
			"backend_ids":        zonal.NewIDStrings(zone, lb.BackendIDs),
			"private_network_id": regional.NewIDString(pnRegion, lb.PrivateNetworkID),
		},
	}
}

func flattenPolicyMetric(metric *autoscaling.Metric) interface{} {
	if metric == nil {
		return nil
	}

	var managedMetric string
	if metric.ManagedMetric != nil {
		managedMetric = metric.ManagedMetric.String()
	}

	return []map[string]interface{}{
		{
			"name":                metric.Name,
			"managed_metric":      managedMetric,
			"cockpit_metric_name": types.FlattenStringPtr(metric.CockpitMetricName),
			"operator":            metric.Operator.String(),
			"aggregate":           metric.Aggregate.String(),
			"sampling_range_min":  metric.SamplingRangeMin,
			"threshold":           metric.Threshold,
		},
	}
}

func flattenVolumes(zone scw.Zone, volMap map[string]*autoscaling.VolumeInstanceTemplate) []interface{} {
	keys := make([]string, 0, len(volMap))
	for k := range volMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	volumes := make([]interface{}, len(keys))

	for i, k := range keys {
		v := volMap[k]
		m := map[string]interface{}{
			"name":        v.Name,
			"boot":        v.Boot,
			"volume_type": v.VolumeType.String(),
			"tags":        v.Tags,
			"perf_iops":   types.FlattenUint32Ptr(v.PerfIops),
		}

		if v.FromEmpty != nil {
			m["from_empty"] = []interface{}{map[string]interface{}{
				"size": int(v.FromEmpty.Size / scw.GB),
			}}
		}

		if v.FromSnapshot != nil {
			inner := map[string]interface{}{
				"snapshot_id": zonal.NewIDString(zone, v.FromSnapshot.SnapshotID),
			}

			if v.FromSnapshot.Size != nil {
				inner["size"] = int(*v.FromSnapshot.Size / scw.GB)
			}

			m["from_snapshot"] = []interface{}{inner}
		}

		volumes[i] = m
	}

	return volumes
}
