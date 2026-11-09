package regional_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func TestLocalityPlanModifier_PlanModifyString(t *testing.T) {
	m := regional.LocalityPlanModifier()

	cases := []struct {
		stateValue    types.String
		planValue     types.String
		wantPlanValue types.String
		name          string
		wantReplace   bool
		wantUnchanged bool
	}{
		{
			// Same UUID, bare in config vs regionalized in state → suppress diff.
			name:          "same UUID with/without region prefix suppresses diff",
			stateValue:    types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
			planValue:     types.StringValue("11111111-1111-1111-1111-111111111111"),
			wantReplace:   false,
			wantPlanValue: types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
		},
		{
			// Different UUID → no suppression (RequiresReplaceIfConfigured handles it).
			name:          "different UUID does not suppress",
			stateValue:    types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
			planValue:     types.StringValue("fr-par/22222222-2222-2222-2222-222222222222"),
			wantReplace:   false,
			wantUnchanged: true,
		},
		{
			// null state → no-op (let other plan modifiers handle).
			name:          "null state is no-op",
			stateValue:    types.StringNull(),
			planValue:     types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
			wantReplace:   false,
			wantUnchanged: true,
		},
		{
			// null plan → no-op.
			name:          "null plan is no-op",
			stateValue:    types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
			planValue:     types.StringNull(),
			wantReplace:   false,
			wantUnchanged: true,
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
			// Same regionalized value on both sides → suppress diff, keep state.
			name:          "identical regional values suppress diff",
			stateValue:    types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
			planValue:     types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
			wantReplace:   false,
			wantPlanValue: types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
		},
		{
			// Different regions, same UUID → suppressed. UUID is the real key
			// identity; the region prefix is just locality context. Two distinct
			// KMS keys always have distinct UUIDs, so this case is theoretical.
			name:          "different region same UUID suppresses diff",
			stateValue:    types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
			planValue:     types.StringValue("nl-ams/11111111-1111-1111-1111-111111111111"),
			wantReplace:   false,
			wantPlanValue: types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
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
			case !tc.wantPlanValue.IsNull():
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
