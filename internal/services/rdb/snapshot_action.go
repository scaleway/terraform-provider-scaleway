package rdb

import (
	"context"
	_ "embed"
	"fmt"
	"time"

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
	_ action.Action              = (*InstanceSnapshotAction)(nil)
	_ action.ActionWithConfigure = (*InstanceSnapshotAction)(nil)
)

// InstanceSnapshotAction creates a snapshot for an RDB instance.
type InstanceSnapshotAction struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

func (a *InstanceSnapshotAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *InstanceSnapshotAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_instance_snapshot"
}

type InstanceSnapshotActionModel struct {
	InstanceID types.String `tfsdk:"instance_id"`
	Region     types.String `tfsdk:"region"`
	Name       types.String `tfsdk:"name"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
	Wait       types.Bool   `tfsdk:"wait"`
}

// NewInstanceSnapshotAction returns a new RDB instance snapshot action.
func NewInstanceSnapshotAction() action.Action {
	return &InstanceSnapshotAction{}
}

//go:embed descriptions/instance_snapshot_action.md
var instanceSnapshotActionDescription string

func (a *InstanceSnapshotAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: instanceSnapshotActionDescription,
		Description:         instanceSnapshotActionDescription,
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "RDB instance ID to snapshot. Can be a plain UUID or a regional ID.",
			},
			"region": regional.SchemaAttribute(),
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the snapshot. If not set, a name will be generated.",
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Description: "Expiration date of the snapshot in RFC3339 format (ISO 8601). If not set, the snapshot will not expire.",
			},
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the snapshot to reach a terminal state before returning.",
			},
		},
	}
}

func (a *InstanceSnapshotAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data InstanceSnapshotActionModel

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

	if data.InstanceID.IsNull() || data.InstanceID.IsUnknown() || data.InstanceID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing instance_id",
			"The instance_id attribute is required to create an RDB snapshot.",
		)

		return
	}

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
		if derivedRegion, id, parseErr := regional.ParseID(data.InstanceID.ValueString()); parseErr == nil {
			region = derivedRegion
			instanceID = id
		} else if a.meta != nil {
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Unable to determine region",
					"Failed to get default region from provider configuration. Please set the region attribute, use a regional instance_id, or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	if region == "" {
		resp.Diagnostics.AddError(
			"Missing region",
			"Could not determine region for RDB snapshot. Please set the region attribute, use a regional instance_id, or configure a default region in the provider.",
		)

		return
	}

	snapshotName := data.Name.ValueString()

	var expirationTime *time.Time

	if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() && data.ExpiresAt.ValueString() != "" {
		expirationRaw := data.ExpiresAt.ValueString()

		parsedTime, err := time.Parse(time.RFC3339, expirationRaw)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid expires_at value",
				fmt.Sprintf("The expires_at attribute must be a valid RFC3339 timestamp. Got %q: %s", expirationRaw, err),
			)

			return
		}

		expirationTime = &parsedTime
	}

	createReq := &rdb.CreateSnapshotRequest{
		Region:     region,
		InstanceID: instanceID,
		ExpiresAt:  expirationTime,
	}

	if snapshotName != "" {
		createReq.Name = snapshotName
	}

	snapshot, err := a.rdbAPI.CreateSnapshot(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing RDB CreateSnapshot action",
			fmt.Sprintf("Failed to create snapshot for instance %s: %s", instanceID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		waitRegion := snapshot.Region
		if waitRegion == "" && region != "" {
			waitRegion = region
		}

		if waitRegion == "" {
			resp.Diagnostics.AddError(
				"Missing region for wait operation",
				"Could not determine region to wait for RDB snapshot completion.",
			)

			return
		}

		_, err = waitForRDBSnapshot(ctx, a.rdbAPI, waitRegion, snapshot.ID, defaultInstanceTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for RDB snapshot completion",
				fmt.Sprintf("Snapshot %s for instance %s did not reach a terminal state: %s", snapshot.ID, instanceID, err),
			)

			return
		}
	}
}
