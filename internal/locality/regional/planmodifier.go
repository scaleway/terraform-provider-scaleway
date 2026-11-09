package regional

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// LocalityPlanModifier suppresses diffs between bare UUIDs and regional IDs
// (e.g. "uuid" vs "fr-par/uuid"). The API returns a bare UUID that the
// provider re-localizes on flatten; without this modifier a bare UUID in
// config would mismatch the regionalized state value after apply.
func LocalityPlanModifier() planmodifier.String {
	return localityPlanModifier{}
}

type localityPlanModifier struct{}

func (m localityPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between bare UUIDs and regional IDs."
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
}
