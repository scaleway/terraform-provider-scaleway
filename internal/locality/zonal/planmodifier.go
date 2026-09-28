package zonal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LocalityPlanModifier suppresses diffs between bare IDs and zonal IDs
// (e.g. "bucket-name" vs "fr-par-1/bucket-name").
//
// During Create, if the plan value lacks a zone prefix, it is normalised
// to include one so the API can resolve the correct zone.
//
// During Update, if the plan and state values differ only by a zone
// prefix, the plan value is carried over unchanged so no diff is
// produced (mimicking the old dsf.Locality behaviour).
func LocalityPlanModifier() planmodifier.String {
	return localityPlanModifier{}
}

type localityPlanModifier struct{}

func (m localityPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between bare IDs and zonal IDs (e.g. \"name\" vs \"fr-par-1/name\")."
}

func (m localityPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m localityPlanModifier) PlanModifyString(
	_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	planStr := req.PlanValue.ValueString()
	expandedPlanID := ExpandID(planStr).ID

	// Add zone prefix if necessary (create path)
	if planStr != expandedPlanID {
		resp.PlanValue = types.StringValue(expandedPlanID)
	}

	// Suppress diffs during update when only the zone prefix differs
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		expandedStateID := ExpandID(req.StateValue.ValueString()).ID

		if expandedStateID == expandedPlanID {
			resp.PlanValue = req.StateValue
		}
	}
}
