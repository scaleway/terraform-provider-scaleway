package regional

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// AllRegions returns all valid Scaleway regions as strings
func AllRegions() []string {
	regions := make([]string, 0, len(scw.AllRegions))
	for _, r := range scw.AllRegions {
		regions = append(regions, r.String())
	}

	return regions
}

// SchemaAttribute returns a Plugin Framework schema attribute for a region field
func SchemaAttribute(description ...string) schema.StringAttribute {
	desc := "The region you want to attach the resource to"
	if len(description) > 0 {
		desc = description[0]
	}

	return schema.StringAttribute{
		Optional:    true,
		Description: desc,
		Validators: []validator.String{
			verify.IsStringOneOfWithWarning(AllRegions()),
		},
	}
}

// SchemaAttributeComputed returns a Plugin Framework schema attribute for a region field,
// with the `Computed` field set to `true`.
func SchemaAttributeComputed(description ...string) schema.StringAttribute {
	desc := "The region you want to attach the resource to"
	if len(description) > 0 {
		desc = description[0]
	}

	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: desc,
		Validators: []validator.String{
			verify.IsStringOneOfWithWarning(AllRegions()),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// LocalityPlanModifier suppresses plan diffs when the configured value differs
// from state only by its locality prefix (e.g. "mybucket" vs "fr-par/mybucket").
type LocalityPlanModifier struct{}

func (LocalityPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs that differ only by the locality prefix."
}

func (LocalityPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Suppresses diffs that differ only by the locality prefix."
}

func (LocalityPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	if locality.ExpandID(req.PlanValue.ValueString()) == locality.ExpandID(req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
