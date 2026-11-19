package jobs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	jobs "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha2"
	types2 "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

type cronConfigSpec struct {
	schedule       string
	timezone       string
	startupCommand []string
	args           []string
}

func expandCronConfig(ctx context.Context, cronConfig types.Object, diags *diag.Diagnostics) cronConfigSpec {
	spec := cronConfigSpec{}

	if cronConfig.IsNull() || cronConfig.IsUnknown() {
		return spec
	}

	cronConfigAttributes := cronConfig.Attributes()

	schedule := cronConfigAttributes["schedule"].(types.String)
	if !schedule.IsNull() && !schedule.IsUnknown() {
		spec.schedule = schedule.ValueString()
	}

	timezone := cronConfigAttributes["timezone"].(types.String)
	if !timezone.IsNull() && !timezone.IsUnknown() {
		spec.timezone = timezone.ValueString()
	}

	startupCommand := cronConfigAttributes["startup_command"].(types.List)
	spec.startupCommand = types2.ExpandUpdatedStringList(ctx, startupCommand, diags)

	args := cronConfigAttributes["args"].(types.List)
	spec.args = types2.ExpandUpdatedStringList(ctx, args, diags)

	return spec
}

func (c cronConfigSpec) ToCreateRequest() *jobs.CreateTriggerRequestCronConfig {
	return &jobs.CreateTriggerRequestCronConfig{
		Schedule:       c.schedule,
		Timezone:       c.timezone,
		StartupCommand: c.startupCommand,
		Args:           c.args,
	}
}

func (c cronConfigSpec) ToUpdateRequest() *jobs.UpdateTriggerRequestCronConfig {
	return &jobs.UpdateTriggerRequestCronConfig{
		Schedule:       &c.schedule,
		Timezone:       &c.timezone,
		StartupCommand: &c.startupCommand,
		Args:           &c.args,
	}
}

func cronConfigAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"schedule": types.StringType,
		"timezone": types.StringType,
		"startup_command": types.ListType{
			ElemType: types.StringType,
		},
		"args": types.ListType{
			ElemType: types.StringType,
		},
	}
}

func flattenCronConfig(ctx context.Context, cron *jobs.TriggerCronConfig, reference any, diags *diag.Diagnostics) types.Object {
	if cron == nil {
		return types.ObjectNull(cronConfigAttributeTypes())
	}

	ccFlat := map[string]attr.Value{
		"schedule": types.StringValue(cron.Schedule),
		"timezone": types.StringValue(cron.Timezone),
	}

	startupCommand, d := types2.FlattenStringList(ctx, "cron_config.startup_command", cron.StartupCommand, reference)
	diags.Append(d...)

	ccFlat["startup_command"] = startupCommand

	args, d := types2.FlattenStringList(ctx, "cron_config.args", cron.Args, reference)
	diags.Append(d...)

	ccFlat["args"] = args

	if diags.HasError() {
		return types.ObjectNull(cronConfigAttributeTypes())
	}

	ccObject, d := types.ObjectValue(cronConfigAttributeTypes(), ccFlat)
	diags.Append(d...)

	return ccObject
}
