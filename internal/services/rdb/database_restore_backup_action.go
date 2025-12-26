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
	_ action.Action              = (*DatabaseBackupRestoreAction)(nil)
	_ action.ActionWithConfigure = (*DatabaseBackupRestoreAction)(nil)
)

// DatabaseBackupRestoreAction restores a database backup.
type DatabaseBackupRestoreAction struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

func (a *DatabaseBackupRestoreAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *DatabaseBackupRestoreAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_database_restore_backup"
}

type DatabaseBackupRestoreActionModel struct {
	BackupID     types.String `tfsdk:"backup_id"`
	InstanceID   types.String `tfsdk:"instance_id"`
	DatabaseName types.String `tfsdk:"database_name"`
	Region       types.String `tfsdk:"region"`
	Wait         types.Bool   `tfsdk:"wait"`
}

// NewDatabaseBackupRestoreAction returns a new RDB database backup restore action.
func NewDatabaseBackupRestoreAction() action.Action {
	return &DatabaseBackupRestoreAction{}
}

//go:embed descriptions/database_backup_restore_action.md
var databaseBackupRestoreDescription string

func (a *DatabaseBackupRestoreAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: databaseBackupRestoreDescription,
		Description:         databaseBackupRestoreDescription,
		Attributes: map[string]schema.Attribute{
			"backup_id": schema.StringAttribute{
				Required:    true,
				Description: "Database backup ID to restore. Can be a plain UUID or a regional ID.",
			},
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "RDB instance ID to restore the backup to. Can be a plain UUID or a regional ID.",
			},
			"database_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the database to restore. If not set, the original database name will be used.",
			},
			"region": regional.SchemaAttribute(),
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the restore operation to complete before returning.",
			},
		},
	}
}

func (a *DatabaseBackupRestoreAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data DatabaseBackupRestoreActionModel

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
			"The backup_id attribute is required to restore a database backup.",
		)

		return
	}

	if data.InstanceID.IsNull() || data.InstanceID.IsUnknown() || data.InstanceID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing instance_id",
			"The instance_id attribute is required to restore a database backup.",
		)

		return
	}

	backupID := locality.ExpandID(data.BackupID.ValueString())
	instanceID := locality.ExpandID(data.InstanceID.ValueString())

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
		} else if derivedRegion, id, parseErr := regional.ParseID(data.InstanceID.ValueString()); parseErr == nil {
			region = derivedRegion
			instanceID = id
		} else if a.meta != nil {
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Unable to determine region",
					"Failed to get default region from provider configuration. Please set the region attribute, use a regional backup_id or instance_id, or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	if region == "" {
		resp.Diagnostics.AddError(
			"Missing region",
			"Could not determine region for RDB database backup restore. Please set the region attribute, use a regional backup_id or instance_id, or configure a default region in the provider.",
		)

		return
	}

	restoreReq := &rdb.RestoreDatabaseBackupRequest{
		Region:           region,
		DatabaseBackupID: backupID,
		InstanceID:       instanceID,
	}

	if !data.DatabaseName.IsNull() && !data.DatabaseName.IsUnknown() && data.DatabaseName.ValueString() != "" {
		restoreReq.DatabaseName = scw.StringPtr(data.DatabaseName.ValueString())
	}

	_, err := a.rdbAPI.RestoreDatabaseBackup(restoreReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing RDB RestoreDatabaseBackup action",
			fmt.Sprintf("Failed to restore backup %s to instance %s: %s", backupID, instanceID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		_, err = waitForRDBInstance(ctx, a.rdbAPI, region, instanceID, defaultInstanceTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for RDB database backup restore completion",
				fmt.Sprintf("Restore operation for instance %s did not complete: %s", instanceID, err),
			)

			return
		}
	}
}
