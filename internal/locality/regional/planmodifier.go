package regional

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// LocalityPlanModifier suppresses diffs between bare IDs and regional IDs
// (e.g. "my-bucket" vs "fr-par/my-bucket"). It mirrors the SDKv2 dsf.Locality
// DiffSuppressFunc: the state value (bare form) is kept so config authors can
// reference either form interchangeably.
//
// Two cases matter, enforced by Terraform Core's plan/apply consistency:
//
//   - Create (no prior state): config may carry a regional ID ("nl-ams/name")
//     but the provider stores the bare name ("name") after apply. Terraform
//     forbids planned != config for a Computed attr without prior state, so
//     we mark the value unknown on create when the config is regional. Apply
//     then accepts the bare value the provider returns.
//   - Update/refresh (prior state present): drift suppression allows a
//     Computed attr to keep the prior state value when config differs, so we
//     keep the (bare) state value whenever the underlying ID matches.
func LocalityPlanModifier() planmodifier.String {
	return localityPlanModifier{}
}

type localityPlanModifier struct{}

func (m localityPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between bare IDs and regional IDs."
}

func (m localityPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m localityPlanModifier) PlanModifyString(
	_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	// Create: no prior state. Leave the plan at the config value.
	if req.StateValue.IsNull() {
		return
	}

	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	// Update/refresh: keep the prior (bare) state value when the underlying
	// ID is unchanged (drift suppression).
	if locality.ExpandID(req.StateValue.ValueString()) == locality.ExpandID(req.PlanValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
