package zonal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// LocalityPlanModifier suppresses diffs between bare UUIDs and zonal IDs
// (e.g. "uuid" vs "fr-par-1/uuid") and forces replacement when the underlying
// UUID actually changes. It also forces replacement when the attribute goes
// from null (not in state) to a value (in config), meaning a snapshot_id is
// being added to an existing volume.
func LocalityPlanModifier() planmodifier.String {
	return localityPlanModifier{}
}

type localityPlanModifier struct{}

func (m localityPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between bare UUIDs and zonal IDs, and forces replacement when the UUID changes."
}

func (m localityPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m localityPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		if !req.PlanValue.IsNull() && req.PlanValue.ValueString() != "" {
			resp.RequiresReplace = true
		}

		return
	}

	if req.PlanValue.IsNull() {
		return
	}

	stateID := locality.ExpandID(req.StateValue.ValueString())
	planID := locality.ExpandID(req.PlanValue.ValueString())

	if stateID == planID {
		resp.PlanValue = req.StateValue

		return
	}

	resp.RequiresReplace = true
}
