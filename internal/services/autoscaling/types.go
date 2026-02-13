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

func expandInstanceCapacity(raw any) *autoscaling.Capacity {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)
	capacity := &autoscaling.Capacity{
		MaxReplicas: uint32(rawMap["max_replicas"].(int)),
		MinReplicas: uint32(rawMap["min_replicas"].(int)),
	}

	if rawVal, ok := rawMap["cooldown_delay"].(int); ok {
		capacity.CooldownDelay = &scw.Duration{Seconds: int64(rawVal)}
	}

	return capacity
}

func expandUpdateInstanceCapacity(raw any) *autoscaling.UpdateInstanceGroupRequestCapacity {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)
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

func expandInstanceLoadBalancer(raw any) *autoscaling.Loadbalancer {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &autoscaling.Loadbalancer{
		ID:               locality.ExpandID(rawMap["id"].(string)),
		PrivateNetworkID: locality.ExpandID(rawMap["private_network_id"].(string)),
		BackendIDs:       locality.ExpandIDs(rawMap["backend_ids"]),
	}
}

func expandPolicyMetric(raw any) *autoscaling.Metric {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

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

func expandUpdatePolicyMetric(raw any) *autoscaling.UpdateInstancePolicyRequestMetric {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

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
		Threshold:         new(float32(rawMap["sampling_threshold"].(int))),
	}
}

func expandUpdateInstanceLoadBalancer(raw any) *autoscaling.UpdateInstanceGroupRequestLoadbalancer {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)
	lb := &autoscaling.UpdateInstanceGroupRequestLoadbalancer{}

	if rawVal, ok := rawMap["backend_ids"].(int); ok {
		lb.BackendIDs = types.ExpandStringsPtr(locality.ExpandIDs(rawVal))
	}

	return lb
}

func expandVolumes(rawVols []any) map[string]*autoscaling.VolumeInstanceTemplate {
	vols := make(map[string]*autoscaling.VolumeInstanceTemplate, len(rawVols))

	for i, raw := range rawVols {
		m := raw.(map[string]any)
		vt := &autoscaling.VolumeInstanceTemplate{
			Name:       m["name"].(string),
			Boot:       m["boot"].(bool),
			VolumeType: autoscaling.VolumeInstanceTemplateVolumeType(m["volume_type"].(string)),
			Tags:       types.ExpandStrings(m["tags"].([]any)),
			PerfIops:   types.ExpandUint32Ptr(m["perf_iops"].(int)),
		}

		if slice := m["from_empty"].([]any); len(slice) > 0 && slice[0] != nil {
			inner := slice[0].(map[string]any)
			sizeGB := inner["size"].(int)
			vt.FromEmpty = &autoscaling.VolumeInstanceTemplateFromEmpty{
				Size: scw.Size(uint64(sizeGB) * uint64(scw.GB)),
			}
		}

		if slice := m["from_snapshot"].([]any); len(slice) > 0 && slice[0] != nil {
			inner := slice[0].(map[string]any)
			snapshot := &autoscaling.VolumeInstanceTemplateFromSnapshot{
				SnapshotID: locality.ExpandID(inner["snapshot_id"].(string)),
			}

			if sz, ok := inner["size"].(int); ok && sz > 0 {
				snapshot.Size = new(scw.Size(uint64(sz) * uint64(scw.GB)))
			}

			vt.FromSnapshot = snapshot
		}

		vols[strconv.Itoa(i)] = vt
	}

	return vols
}

func flattenInstanceCapacity(capacity *autoscaling.Capacity) any {
	if capacity == nil {
		return nil
	}

	return []map[string]any{
		{
			"max_replicas":   capacity.MaxReplicas,
			"min_replicas":   capacity.MinReplicas,
			"cooldown_delay": capacity.CooldownDelay.Seconds,
		},
	}
}

func flattenInstanceLoadBalancer(lb *autoscaling.Loadbalancer, zone scw.Zone) any {
	if lb == nil {
		return nil
	}

	pnRegion, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	return []map[string]any{
		{
			"id":                 zonal.NewIDString(zone, lb.ID),
			"backend_ids":        zonal.NewIDStrings(zone, lb.BackendIDs),
			"private_network_id": regional.NewIDString(pnRegion, lb.PrivateNetworkID),
		},
	}
}

func flattenPolicyMetric(metric *autoscaling.Metric) any {
	if metric == nil {
		return nil
	}

	var managedMetric string
	if metric.ManagedMetric != nil {
		managedMetric = metric.ManagedMetric.String()
	}

	return []map[string]any{
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

func flattenVolumes(zone scw.Zone, volMap map[string]*autoscaling.VolumeInstanceTemplate) []any {
	keys := make([]string, 0, len(volMap))
	for k := range volMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	volumes := make([]any, len(keys))

	for i, k := range keys {
		v := volMap[k]
		m := map[string]any{
			"name":        v.Name,
			"boot":        v.Boot,
			"volume_type": v.VolumeType.String(),
			"tags":        v.Tags,
			"perf_iops":   types.FlattenUint32Ptr(v.PerfIops),
		}

		if v.FromEmpty != nil {
			m["from_empty"] = []any{map[string]any{
				"size": int(v.FromEmpty.Size / scw.GB),
			}}
		}

		if v.FromSnapshot != nil {
			inner := map[string]any{
				"snapshot_id": zonal.NewIDString(zone, v.FromSnapshot.SnapshotID),
			}

			if v.FromSnapshot.Size != nil {
				inner["size"] = int(*v.FromSnapshot.Size / scw.GB)
			}

			m["from_snapshot"] = []any{inner}
		}

		volumes[i] = m
	}

	return volumes
}
