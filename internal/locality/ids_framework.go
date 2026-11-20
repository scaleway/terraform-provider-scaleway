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

// frameworkCollection is implemented by both types.List and types.Set
type frameworkCollection interface {
	IsNull() bool
	IsUnknown() bool
	ElementsAs(ctx context.Context, target any, allowUnhandled bool) diag.Diagnostics
}

func ExpandFrameworkIDs(ctx context.Context, val frameworkCollection) ([]string, diag.Diagnostics) {
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
