package list

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TagsModel interface {
	GetTags() types.List
}

func ExtractTags(ctx context.Context, data TagsModel) ([]string, diag.Diagnostics) {
	var tags []string
	tagsList := data.GetTags()
	if !tagsList.IsNull() {
		diags := tagsList.ElementsAs(ctx, &tags, false)
		if diags.HasError() {
			return nil, diags
		}
	}
	return tags, nil
}
