package list

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FileSystemIDs interface {
	GetFileSystemIDs() types.List
}

func ExtractFileSystemIDs(ctx context.Context, data FileSystemIDs) ([]string, diag.Diagnostics) {
	var fsIDs []string

	fsIDsList := data.GetFileSystemIDs()
	if !fsIDsList.IsNull() {
		diags := fsIDsList.ElementsAs(ctx, &fsIDs, false)
		if diags.HasError() {
			return nil, diags
		}
	}

	return fsIDs, nil
}
