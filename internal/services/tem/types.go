package tem

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

type domainReputationModel struct {
	Status           types.String `tfsdk:"status"`
	Score            types.Int64  `tfsdk:"score"`
	ScoredAt         types.String `tfsdk:"scored_at"`
	PreviousScore    types.Int64  `tfsdk:"previous_score"`
	PreviousScoredAt types.String `tfsdk:"previous_scored_at"`
}

func domainReputationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":             types.StringType,
		"score":              types.Int64Type,
		"scored_at":          types.StringType,
		"previous_score":     types.Int64Type,
		"previous_scored_at": types.StringType,
	}
}

func flattenDomainReputation(ctx context.Context, reputation *tem.DomainReputation, diags *diag.Diagnostics) types.List {
	objType := types.ObjectType{AttrTypes: domainReputationAttrTypes()}
	if reputation == nil {
		return types.ListNull(objType)
	}

	previousScore := int64(0)
	if reputation.PreviousScore != nil {
		previousScore = int64(*reputation.PreviousScore)
	}

	models := []domainReputationModel{{
		Status:           types.StringValue(reputation.Status.String()),
		Score:            types.Int64Value(int64(reputation.Score)),
		ScoredAt:         types.StringValue(scwtypes.FlattenTime(reputation.ScoredAt).(string)),
		PreviousScore:    types.Int64Value(previousScore),
		PreviousScoredAt: types.StringValue(scwtypes.FlattenTime(reputation.PreviousScoredAt).(string)),
	}}

	list, d := types.ListValueFrom(ctx, objType, models)
	diags.Append(d...)

	return list
}

func expandWebhookEventTypes(ctx context.Context, list types.List, diags *diag.Diagnostics) []tem.WebhookEventType {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var values []string

	diags.Append(list.ElementsAs(ctx, &values, false)...)

	if diags.HasError() {
		return nil
	}

	eventTypes := make([]tem.WebhookEventType, len(values))
	for i, v := range values {
		eventTypes[i] = tem.WebhookEventType(v)
	}

	return eventTypes
}

func flattenWebhookEventTypes(ctx context.Context, eventTypes []tem.WebhookEventType, diags *diag.Diagnostics) types.List {
	values := make([]string, len(eventTypes))
	for i, v := range eventTypes {
		values[i] = string(v)
	}

	list, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)

	return list
}
