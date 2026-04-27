package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FlattenFrameworkStringValue(s string) types.String {
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}

func FlattenFrameworkStringList(ctx context.Context, items []string) (types.List, diag.Diagnostics) {
	if len(items) == 0 {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValueFrom(ctx, types.StringType, items)
}
