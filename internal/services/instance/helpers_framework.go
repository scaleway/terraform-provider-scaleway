package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// expandRawIDList builds an array of raw UUIDs in string from (without locality) from a Terraform types.List
func expandRawIDList(ctx context.Context, list types.List, attributeName string, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)

	rawIDs := make([]string, 0, len(result))
	for _, id := range result {
		rawID, err := locality.ExtractUUID(id)
		if err != nil {
			diags.AddAttributeError(path.Root(attributeName), "Failed to parse "+attributeName, err.Error())

			return nil
		}

		rawIDs = append(rawIDs, rawID)
	}

	return rawIDs
}
