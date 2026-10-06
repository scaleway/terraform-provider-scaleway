package autoscaling

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	autoscalingAlpha1 "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	autoscalingAlpha2 "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// NewAPIWithZone returns a new autoscaling API and the zone for a Create request
func NewAPIWithZone(d *schema.ResourceData, m any) (*autoscalingAlpha1.API, scw.Zone, error) {
	autoscalingAPI := autoscalingAlpha1.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return autoscalingAPI, zone, nil
}

// NewAPIWithZoneAndID returns a new autoscaling API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, zonalID string) (*autoscalingAlpha1.API, scw.Zone, string, error) {
	autoscalingAPI := autoscalingAlpha1.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonalID)
	if err != nil {
		return nil, "", "", err
	}

	return autoscalingAPI, zone, ID, nil
}

func expandScalingPolicy(policy types.Object) *autoscalingAlpha2.ScalingPolicySpec {
	spec := &autoscalingAlpha2.ScalingPolicySpec{}

	if policy.IsNull() || policy.IsUnknown() {
		return spec
	}

	policyAttributes := policy.Attributes()

	minSize := policyAttributes["minimum_size"].(types.Int32)
	if !minSize.IsNull() && !minSize.IsUnknown() {
		spec.MinimumSize = new(uint32(minSize.ValueInt32()))
	}

	maxSize := policyAttributes["maximum_size"].(types.Int32)
	if !maxSize.IsNull() && !maxSize.IsUnknown() {
		spec.MaximumSize = new(uint32(maxSize.ValueInt32()))
	}

	scaleInCooldown := policyAttributes["scale_in_cooldown"].(timetypes.GoDuration)
	if !scaleInCooldown.IsNull() && !scaleInCooldown.IsUnknown() {
		duration, _ := time.ParseDuration(scaleInCooldown.ValueString())
		spec.ScaleInCooldown = scw.NewDurationFromTimeDuration(duration)
	}

	scaleOutCooldown := policyAttributes["scale_out_cooldown"].(timetypes.GoDuration)
	if !scaleOutCooldown.IsNull() && !scaleOutCooldown.IsUnknown() {
		duration, _ := time.ParseDuration(scaleOutCooldown.ValueString())
		spec.ScaleOutCooldown = scw.NewDurationFromTimeDuration(duration)
	}

	scaleInStep := policyAttributes["scale_in_step"].(types.Int32)
	if !scaleInStep.IsNull() && !scaleInStep.IsUnknown() {
		spec.ScaleInStep = new(uint32(scaleInStep.ValueInt32()))
	}

	scaleOutStep := policyAttributes["scale_out_step"].(types.Int32)
	if !scaleOutStep.IsNull() && !scaleOutStep.IsUnknown() {
		spec.ScaleOutStep = new(uint32(scaleOutStep.ValueInt32()))
	}

	fixedSize := policyAttributes["fixed_size"].(types.Int32)
	if !fixedSize.IsNull() && !fixedSize.IsUnknown() {
		spec.FixedSize = &autoscalingAlpha2.GroupScalingPolicyScalingPolicyFixedSize{
			Size: uint32(fixedSize.ValueInt32()),
		}
	}

	cpuTarget := policyAttributes["cpu_target"].(types.Int32)
	if !cpuTarget.IsNull() && !cpuTarget.IsUnknown() {
		spec.CPUTarget = &autoscalingAlpha2.GroupScalingPolicyScalingPolicyCPUTarget{
			TargetAvgPercent: uint32(cpuTarget.ValueInt32()),
		}
	}

	memoryTarget := policyAttributes["memory_target"].(types.Int32)
	if !memoryTarget.IsNull() && !memoryTarget.IsUnknown() {
		spec.MemoryTarget = &autoscalingAlpha2.GroupScalingPolicyScalingPolicyMemoryTarget{
			TargetAvgPercent: uint32(memoryTarget.ValueInt32()),
		}
	}

	return spec
}

func scalingPolicyAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"minimum_size":       types.Int32Type,
		"maximum_size":       types.Int32Type,
		"scale_in_cooldown":  timetypes.GoDurationType{},
		"scale_out_cooldown": timetypes.GoDurationType{},
		"scale_in_step":      types.Int32Type,
		"scale_out_step":     types.Int32Type,
		"fixed_size":         types.Int32Type,
		"cpu_target":         types.Int32Type,
		"memory_target":      types.Int32Type,
	}
}

func flattenScalingPolicy(policy *autoscalingAlpha2.GroupScalingPolicy) (types.Object, diag.Diagnostics) {
	if policy == nil {
		return types.ObjectNull(scalingPolicyAttributeTypes()), nil
	}

	spFlat := map[string]attr.Value{
		"minimum_size":       types.Int32Value(int32(policy.MinimumSize)),
		"maximum_size":       types.Int32Value(int32(policy.MaximumSize)),
		"scale_in_cooldown":  timetypes.NewGoDurationNull(),
		"scale_out_cooldown": timetypes.NewGoDurationNull(),
		"scale_in_step":      types.Int32Value(int32(policy.ScaleInStep)),
		"scale_out_step":     types.Int32Value(int32(policy.ScaleOutStep)),
		"fixed_size":         types.Int32Null(),
		"cpu_target":         types.Int32Null(),
		"memory_target":      types.Int32Null(),
	}

	if policy.ScaleInCooldown != nil {
		spFlat["scale_in_cooldown"] = timetypes.NewGoDurationValue(*policy.ScaleInCooldown.ToTimeDuration())
	}

	if policy.ScaleOutCooldown != nil {
		spFlat["scale_out_cooldown"] = timetypes.NewGoDurationValue(*policy.ScaleOutCooldown.ToTimeDuration())
	}

	if policy.FixedSize != nil {
		spFlat["fixed_size"] = types.Int32Value(int32(policy.FixedSize.Size))
	}

	if policy.CPUTarget != nil {
		spFlat["cpu_target"] = types.Int32Value(int32(policy.CPUTarget.TargetAvgPercent))
	}

	if policy.MemoryTarget != nil {
		spFlat["memory_target"] = types.Int32Value(int32(policy.MemoryTarget.TargetAvgPercent))
	}

	return types.ObjectValue(scalingPolicyAttributeTypes(), spFlat)
}

func expandLoadBalancerConfiguration(ctx context.Context, lbConfig types.Object, d *diag.Diagnostics) *autoscalingAlpha2.LoadBalancerConfigurationSpec {
	lbConfigSpec := &autoscalingAlpha2.LoadBalancerConfigurationSpec{}

	if lbConfig.IsNull() || lbConfig.IsUnknown() {
		return nil
	}

	configAttributes := lbConfig.Attributes()

	loadBalancerID := configAttributes["load_balancer_id"].(types.String)
	if !loadBalancerID.IsNull() && !loadBalancerID.IsUnknown() {
		lbConfigSpec.LoadBalancerID = scwtypes.ExpandRawID(loadBalancerID, "load_balancer_configuration.load_balancer_id", d)
	}

	backends := configAttributes["backends"].(types.List)
	if !backends.IsNull() && !backends.IsUnknown() {
		lbConfigSpec.Backends = expandLBConfigBackends(ctx, backends, d)
	}

	autohealing := configAttributes["auto_healing"].(types.Object)
	if !autohealing.IsNull() && !autohealing.IsUnknown() {
		lbConfigSpec.AutoHealing = expandLBConfigAutohealing(autohealing)
	}

	return lbConfigSpec
}

func expandLBConfigBackends(ctx context.Context, backendsList types.List, diags *diag.Diagnostics) []*autoscalingAlpha2.LoadBalancerConfigurationSpecBackend {
	backendCount := backendsList.Length(basetypes.CollectionLengthOptions{})
	backendSpecs := make([]*autoscalingAlpha2.LoadBalancerConfigurationSpecBackend, 0, backendCount)

	backendObjects := make([]types.Object, 0, backendCount)
	diags.Append(backendsList.ElementsAs(ctx, &backendObjects, false)...)

	if diags.HasError() {
		return backendSpecs
	}

	for _, backendObject := range backendObjects {
		backend := &autoscalingAlpha2.LoadBalancerConfigurationSpecBackend{}
		backendObjAttrs := backendObject.Attributes()

		backendID := backendObjAttrs["backend_id"].(types.String)
		backend.BackendID = *scwtypes.ExpandRawID(backendID, "backend_id", diags)

		addressFamily := backendObjAttrs["address_family"].(types.String)
		backend.AddressFamily = autoscalingAlpha2.GroupLoadBalancerConfigurationBackendAddressFamily(addressFamily.ValueString())

		privateNetworkID := backendObjAttrs["private_network_id"].(types.String)
		if !privateNetworkID.IsNull() && !privateNetworkID.IsUnknown() {
			backend.PrivateNetworkID = scwtypes.ExpandRawID(privateNetworkID, "private_network_id", diags)
		}

		backendSpecs = append(backendSpecs, backend)
	}

	return backendSpecs
}

func expandLBConfigAutohealing(autohealingObject types.Object) *autoscalingAlpha2.LoadBalancerConfigurationSpecAutoHealing {
	autoHealingSpec := &autoscalingAlpha2.LoadBalancerConfigurationSpecAutoHealing{}

	enabledAttr := autohealingObject.Attributes()["enabled"].(types.Bool)
	if !enabledAttr.IsNull() && !enabledAttr.IsUnknown() {
		autoHealingSpec.Enabled = new(enabledAttr.ValueBool())
	}

	gracePeriod := autohealingObject.Attributes()["grace_period"].(timetypes.GoDuration)
	if !gracePeriod.IsNull() && !gracePeriod.IsUnknown() {
		gracePeriodDuration, _ := time.ParseDuration(gracePeriod.ValueString())
		autoHealingSpec.GracePeriod = scw.NewDurationFromTimeDuration(gracePeriodDuration)
	}

	return autoHealingSpec
}

func loadBalancerConfigAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"load_balancer_id": types.StringType,
		"backends": types.ListType{
			ElemType: backendObjectType(),
		},
		"auto_healing": autoHealingObjectType(),
	}
}

func flattenLoadBalancerConfiguration(ctx context.Context, config *autoscalingAlpha2.GroupLoadBalancerConfiguration, zone scw.Zone, reference any) (types.Object, diag.Diagnostics) {
	if config == nil {
		return types.ObjectNull(loadBalancerConfigAttributeTypes()), nil
	}

	diags := diag.Diagnostics{}
	lbConfFlat := map[string]attr.Value{
		"backends":     flattenBackends(ctx, config.Backends, zone, reference, &diags),
		"auto_healing": flattenAutoHealing(config.AutoHealing, &diags),
	}

	if scwtypes.IDUsesZonedFormat(ctx, reference, path.Root("load_balancer_configuration").AtName("load_balancer_id"), &diags) {
		lbConfFlat["load_balancer_id"] = types.StringValue(zonal.NewIDString(zone, config.LoadBalancerID))
	} else {
		lbConfFlat["load_balancer_id"] = types.StringValue(config.LoadBalancerID)
	}

	if diags.HasError() {
		return types.ObjectNull(loadBalancerConfigAttributeTypes()), diags
	}

	return types.ObjectValue(loadBalancerConfigAttributeTypes(), lbConfFlat)
}

func flattenBackends(ctx context.Context, backends []*autoscalingAlpha2.GroupLoadBalancerConfigurationBackend, zone scw.Zone, reference any, diags *diag.Diagnostics) types.List {
	if len(backends) == 0 {
		return types.ListNull(backendObjectType())
	}

	backendObjects := make([]attr.Value, 0, len(backends))
	for index, backend := range backends {
		backendAttrs := map[string]attr.Value{
			"address_family": types.StringValue(backend.AddressFamily.String()),
		}

		if scwtypes.IDUsesZonedFormat(ctx, reference, path.Root("load_balancer_configuration").AtName("backends").AtListIndex(index).AtName("backend_id"), diags) {
			backendAttrs["backend_id"] = types.StringValue(zonal.NewIDString(zone, backend.BackendID))
		} else {
			backendAttrs["backend_id"] = types.StringValue(backend.BackendID)
		}

		if backend.PrivateNetworkID != nil {
			if scwtypes.IDUsesRegionalFormat(ctx, reference, path.Root("load_balancer_configuration").AtName("backends").AtListIndex(index).AtName("private_network_id"), diags) {
				region, err := zone.Region()
				if err != nil {
					diags.Append(diag.NewErrorDiagnostic(fmt.Sprintf("failed to infer region from zone %q", zone), err.Error()))

					return types.ListNull(backendObjectType())
				}

				backendAttrs["private_network_id"] = types.StringValue(regional.NewIDString(region, *backend.PrivateNetworkID))
			} else {
				backendAttrs["private_network_id"] = types.StringValue(*backend.PrivateNetworkID)
			}
		} else {
			backendAttrs["private_network_id"] = types.StringNull()
		}

		backendObject, d := types.ObjectValue(backendObjectType().AttrTypes, backendAttrs)
		diags.Append(d...)

		if d.HasError() {
			return types.ListNull(backendObjectType())
		}

		backendObjects = append(backendObjects, backendObject)
	}

	backendList, d := types.ListValueFrom(ctx, backendObjectType(), backendObjects)
	diags.Append(d...)

	return backendList
}

func flattenAutoHealing(autoHealing *autoscalingAlpha2.GroupLoadBalancerConfigurationAutoHealing, diags *diag.Diagnostics) types.Object {
	if autoHealing == nil {
		return types.ObjectNull(autoHealingObjectType().AttrTypes)
	}

	autoHealingAttrs := map[string]attr.Value{
		"enabled": types.BoolValue(autoHealing.Enabled),
	}

	if autoHealing.GracePeriod != nil {
		autoHealingAttrs["grace_period"] = timetypes.NewGoDurationValue(*autoHealing.GracePeriod.ToTimeDuration())
	} else {
		autoHealingAttrs["grace_period"] = timetypes.NewGoDurationNull()
	}

	autoHealingObject, d := types.ObjectValue(autoHealingObjectType().AttrTypes, autoHealingAttrs)
	diags.Append(d...)

	return autoHealingObject
}
