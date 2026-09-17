package zonal

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalityPlanModifier_SuppressDiffSameUUID(t *testing.T) {
	m := LocalityPlanModifier()

	req := planmodifier.StringRequest{
		StateValue: types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
		PlanValue:  types.StringValue("11111111-1111-1111-1111-111111111111"),
	}
	resp := &planmodifier.StringResponse{}

	m.PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if resp.RequiresReplace {
		t.Error("expected no replacement when UUIDs match after expansion")
	}
	if resp.PlanValue.ValueString() != req.StateValue.ValueString() {
		t.Errorf("expected plan value to be state value %q, got %q",
			req.StateValue.ValueString(), resp.PlanValue.ValueString())
	}
}

func TestLocalityPlanModifier_DifferentUUIDForcesReplace(t *testing.T) {
	m := LocalityPlanModifier()

	req := planmodifier.StringRequest{
		StateValue: types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
		PlanValue:  types.StringValue("22222222-2222-2222-2222-222222222222"),
	}
	resp := &planmodifier.StringResponse{}

	m.PlanModifyString(context.Background(), req, resp)

	if !resp.RequiresReplace {
		t.Error("expected replacement when UUIDs differ")
	}
}

func TestLocalityPlanModifier_NullStateToValueForcesReplace(t *testing.T) {
	m := LocalityPlanModifier()

	req := planmodifier.StringRequest{
		StateValue: types.StringNull(),
		PlanValue:  types.StringValue("11111111-1111-1111-1111-111111111111"),
	}
	resp := &planmodifier.StringResponse{}

	m.PlanModifyString(context.Background(), req, resp)

	if !resp.RequiresReplace {
		t.Error("expected replacement when going from null state to a value")
	}
}

func TestLocalityPlanModifier_ValueToNullNoReplace(t *testing.T) {
	m := LocalityPlanModifier()

	req := planmodifier.StringRequest{
		StateValue: types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
		PlanValue:  types.StringNull(),
	}
	resp := &planmodifier.StringResponse{}

	m.PlanModifyString(context.Background(), req, resp)

	if resp.RequiresReplace {
		t.Error("expected no replacement when removing from config")
	}
}

func TestLocalityPlanModifier_BothNullNoReplace(t *testing.T) {
	m := LocalityPlanModifier()

	req := planmodifier.StringRequest{
		StateValue: types.StringNull(),
		PlanValue:  types.StringNull(),
	}
	resp := &planmodifier.StringResponse{}

	m.PlanModifyString(context.Background(), req, resp)

	if resp.RequiresReplace {
		t.Error("expected no replacement when both values are null")
	}
}
