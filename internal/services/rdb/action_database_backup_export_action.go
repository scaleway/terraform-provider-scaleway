package rdb

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*DatabaseBackupExportAction)(nil)
	_ action.ActionWithConfigure = (*DatabaseBackupExportAction)(nil)
)

// DatabaseBackupExportAction exports a database backup.
type DatabaseBackupExportAction struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

func (a *DatabaseBackupExportAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.meta = m
	a.rdbAPI = newAPI(m)
}

func (a *DatabaseBackupExportAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_database_backup_export_action"
}

type DatabaseBackupExportActionModel struct {
	BackupID types.String `tfsdk:"backup_id"`
	Region   types.String `tfsdk:"region"`
	Wait     types.Bool   `tfsdk:"wait"`
}

// NewDatabaseBackupExportAction returns a new RDB database backup export action.
func NewDatabaseBackupExportAction() action.Action {
	return &DatabaseBackupExportAction{}
}

//go:embed descriptions/database_backup_export_action.md
var databaseBackupExportDescription string

func (a *DatabaseBackupExportAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: databaseBackupExportDescription,
		Description:         databaseBackupExportDescription,
		Attributes: map[string]schema.Attribute{
			"backup_id": schema.StringAttribute{
				Required:    true,
				Description: "Database backup ID to export. Can be a plain UUID or a regional ID.",
			},
			"region": regional.SchemaAttribute(),
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the export operation to complete before returning.",
			},
		},
	}
}

func (a *DatabaseBackupExportAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data DatabaseBackupExportActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.rdbAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured rdbAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.BackupID.IsNull() || data.BackupID.IsUnknown() || data.BackupID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing backup_id",
			"The backup_id attribute is required to export a database backup.",
		)

		return
	}

	backupID := locality.ExpandID(data.BackupID.ValueString())

	var region scw.Region

	if !data.Region.IsNull() && !data.Region.IsUnknown() && data.Region.ValueString() != "" {
		parsedRegion, err := scw.ParseRegion(data.Region.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid region value",
				fmt.Sprintf("The region attribute must be a valid Scaleway region. Got %q: %s", data.Region.ValueString(), err),
			)

			return
		}

		region = parsedRegion
	} else {
		if derivedRegion, id, parseErr := regional.ParseID(data.BackupID.ValueString()); parseErr == nil {
			region = derivedRegion
			backupID = id
		} else if a.meta != nil {
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Unable to determine region",
					"Failed to get default region from provider configuration. Please set the region attribute, use a regional backup_id, or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	if region == "" {
		resp.Diagnostics.AddError(
			"Missing region",
			"Could not determine region for RDB database backup export. Please set the region attribute, use a regional backup_id, or configure a default region in the provider.",
		)

		return
	}

	exportReq := &rdb.ExportDatabaseBackupRequest{
		Region:           region,
		DatabaseBackupID: backupID,
	}

	_, err := a.rdbAPI.ExportDatabaseBackup(exportReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing RDB ExportDatabaseBackup action",
			fmt.Sprintf("Failed to export backup %s: %s", backupID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		_, err = waitForRDBDatabaseBackup(ctx, a.rdbAPI, region, backupID, defaultInstanceTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for RDB database backup export completion",
				fmt.Sprintf("Export operation for backup %s did not complete: %s", backupID, err),
			)

			return
		}
	}
}
