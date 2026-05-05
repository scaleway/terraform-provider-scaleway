package locality

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ExpandFrameworkID(val types.String) *string {
	if val.IsNull() || val.IsUnknown() {
		return nil
	}

	return new(ExpandID(val.ValueString()))
}

func ExpandFrameworkIDs(ctx context.Context, val types.List) ([]string, diag.Diagnostics) {
	if val.IsNull() || val.IsUnknown() {
		return nil, nil
	}

	var raw []string

	diags := val.ElementsAs(ctx, &raw, false)
	if diags.HasError() {
		return nil, diags
	}

	expanded := make([]string, len(raw))
	for i, id := range raw {
		expanded[i] = ExpandID(id)
	}

	return expanded, nil
}
