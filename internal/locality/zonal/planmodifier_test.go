package zonal_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func TestLocalityPlanModifier_PlanModifyString(t *testing.T) {
	m := zonal.LocalityPlanModifier()

	cases := []struct {
		stateValue    types.String
		planValue     types.String
		wantPlanValue types.String
		name          string
		wantReplace   bool
		wantUnchanged bool
	}{
		{
			// Same UUID, different locality prefix → suppress diff, keep state value.
			name:          "same UUID with/without zone prefix suppresses diff",
			stateValue:    types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
			planValue:     types.StringValue("11111111-1111-1111-1111-111111111111"),
			wantReplace:   false,
			wantPlanValue: types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
		},
		{
			// Different UUID → force replacement.
			name:        "different UUID forces replacement",
			stateValue:  types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
			planValue:   types.StringValue("fr-par-1/22222222-2222-2222-2222-222222222222"),
			wantReplace: true,
		},
		{
			// null → value (adding snapshot_id) → force replacement.
			name:        "null to value forces replacement",
			stateValue:  types.StringNull(),
			planValue:   types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
			wantReplace: true,
		},
		{
			// value → null (removing snapshot_id) → force replacement.
			name:        "value to null forces replacement",
			stateValue:  types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
			planValue:   types.StringNull(),
			wantReplace: true,
		},
		{
			// Both null → no-op.
			name:          "both null is a no-op",
			stateValue:    types.StringNull(),
			planValue:     types.StringNull(),
			wantReplace:   false,
			wantUnchanged: true,
		},
		{
			// Same zoned value on both sides → suppress diff, keep state value.
			name:          "identical zoned values suppress diff",
			stateValue:    types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
			planValue:     types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
			wantReplace:   false,
			wantPlanValue: types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				StateValue: tc.stateValue,
				PlanValue:  tc.planValue,
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tc.planValue,
			}

			m.PlanModifyString(context.Background(), req, resp)

			if resp.RequiresReplace != tc.wantReplace {
				t.Errorf("RequiresReplace = %v, want %v", resp.RequiresReplace, tc.wantReplace)
			}

			if resp.Diagnostics.HasError() {
				t.Errorf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			switch {
			case tc.wantPlanValue.Equal(types.StringNull()) && !tc.wantPlanValue.IsNull():
				// wantPlanValue set (non-null) → assert exact match.
				if resp.PlanValue != tc.wantPlanValue {
					t.Errorf("PlanValue = %q, want %q", resp.PlanValue, tc.wantPlanValue)
				}
			case tc.wantUnchanged:
				if resp.PlanValue != tc.planValue {
					t.Errorf("PlanValue = %q, want unchanged %q", resp.PlanValue, tc.planValue)
				}
			}
		})
	}
}
