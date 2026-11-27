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

// LocalitySetPlanModifier suppresses diffs on a set of IDs when both sides
// reference the same UUIDs, regardless of the region prefix
func LocalitySetPlanModifier() planmodifier.Set {
	return localitySetPlanModifier{}
}

type localitySetPlanModifier struct{}

func (m localitySetPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between sets of bare UUIDs and regional IDs."
}

func (m localitySetPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m localitySetPlanModifier) PlanModifySet(
	ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse,
) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	stateIDs, diags := locality.ExpandFrameworkIDs(ctx, req.StateValue)
	if diags.HasError() {
		return
	}

	// Fails when some elements are still unknown, in which case there is nothing to suppress
	planIDs, diags := locality.ExpandFrameworkIDs(ctx, req.PlanValue)
	if diags.HasError() {
		return
	}

	if len(stateIDs) != len(planIDs) {
		return
	}

	inState := make(map[string]struct{}, len(stateIDs))
	for _, id := range stateIDs {
		inState[id] = struct{}{}
	}

	for _, id := range planIDs {
		if _, ok := inState[id]; !ok {
			return
		}
	}

	resp.PlanValue = req.StateValue
}
