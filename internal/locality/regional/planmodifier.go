package regional

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// LocalityPlanModifier suppresses diffs between bare IDs and regional IDs
// (e.g. "bucket-name" vs "fr-par/bucket-name").
//
// During Update, if the plan and state values differ only by a regional
// prefix, the plan value is normalised to the state value so no diff is
// produced.
func LocalityPlanModifier() planmodifier.String {
	return localityPlanModifier{}
}

type localityPlanModifier struct{}

func (m localityPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between bare IDs and regional IDs (e.g. \"name\" vs \"fr-par/name\")."
}

func (m localityPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m localityPlanModifier) PlanModifyString(
	_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	planStr := req.PlanValue.ValueString()
	expandedPlanID := locality.ExpandID(planStr)

	// Add region prefix if necessary
	if planStr != expandedPlanID {
		resp.PlanValue = types.StringValue(expandedPlanID)
	}

	// 3. Handle Updates (Semantic Equality)
	// If there is a prior state (update phase) and the normalized versions match,
	// suppress the diff by carrying over the exact state value.
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		stateID := locality.ExpandID(req.StateValue.ValueString())

		if stateID == expandedPlanID {
			resp.PlanValue = req.StateValue
		}
	}
}
