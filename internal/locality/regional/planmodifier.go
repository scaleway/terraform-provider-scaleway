package regional

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// LocalityPlanModifier suppresses diffs between bare UUIDs and regional IDs
// (e.g. "uuid" vs "fr-par/uuid").
func LocalityPlanModifier() planmodifier.String {
	return localityPlanModifier{}
}

type localityPlanModifier struct{}

func (m localityPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between bare UUIDs and regional IDs, " +
		"and forces replacement when the region changes."
}

func (m localityPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m localityPlanModifier) PlanModifyString(
	_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	if req.StateValue.IsNull() {
		return
	}

	if req.PlanValue.IsNull() {
		return
	}

	stateID := locality.ExpandID(req.StateValue.ValueString())
	planID := locality.ExpandID(req.PlanValue.ValueString())

	if stateID == planID {
		resp.PlanValue = req.StateValue
	}

	// Force replacement when both sides have regions and regions differ
	stateHasRegion := strings.Contains(req.StateValue.ValueString(), "/")
	planHasRegion := strings.Contains(req.PlanValue.ValueString(), "/")

	if stateHasRegion && planHasRegion {
		stateRegion := ExpandID(req.StateValue.ValueString()).Region
		planRegion := ExpandID(req.PlanValue.ValueString()).Region

		if stateRegion != planRegion {
			resp.RequiresReplace = true
		}
	}
}
