package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*InstanceSnapshotAction)(nil)
	_ action.ActionWithConfigure = (*InstanceSnapshotAction)(nil)
)

// InstanceSnapshotAction creates a snapshot for a MongoDB instance.
type InstanceSnapshotAction struct {
	mongodbAPI *mongodb.API
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

	a.mongodbAPI = newAPI(m)
}

func (a *InstanceSnapshotAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mongodb_instance_snapshot_action"
}

type InstanceSnapshotActionModel struct {
	InstanceID types.String `tfsdk:"instance_id"`
	Region     types.String `tfsdk:"region"`
	Name       types.String `tfsdk:"name"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
	Wait       types.Bool   `tfsdk:"wait"`
}

// NewInstanceSnapshotAction returns a new MongoDB instance snapshot action.
func NewInstanceSnapshotAction() action.Action {
	return &InstanceSnapshotAction{}
}

func (a *InstanceSnapshotAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "MongoDB instance ID to snapshot. Can be a plain UUID or a regional ID.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region of the MongoDB instance. If not set, the region is derived from the instance_id when possible or from the provider configuration.",
			},
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

	if a.mongodbAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured mongodbAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.InstanceID.IsNull() || data.InstanceID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing instance_id",
			"The instance_id attribute is required to create a MongoDB snapshot.",
		)

		return
	}

	instanceID := locality.ExpandID(data.InstanceID.ValueString())

	var (
		region scw.Region
		err    error
	)

	if !data.Region.IsNull() && data.Region.ValueString() != "" {
		region = scw.Region(data.Region.ValueString())
	} else {
		// Try to derive region from the instance_id if it is a regional ID.
		if derivedRegion, id, parseErr := regional.ParseID(data.InstanceID.ValueString()); parseErr == nil {
			region = derivedRegion
			instanceID = id
		}
	}

	snapshotName := data.Name.ValueString()
	if snapshotName == "" {
		snapshotName = "tf-mongodb-snapshot"
	}

	var expirationTime *time.Time
	if !data.ExpiresAt.IsNull() && data.ExpiresAt.ValueString() != "" {
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

	createReq := &mongodb.CreateSnapshotRequest{
		InstanceID: instanceID,
		Name:       snapshotName,
		ExpiresAt:  expirationTime,
	}

	if region != "" {
		createReq.Region = region
	}

	snapshot, err := a.mongodbAPI.CreateSnapshot(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing MongoDB CreateSnapshot action",
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
				"Could not determine region to wait for MongoDB snapshot completion.",
			)

			return
		}

		_, err = waitForSnapshot(ctx, a.mongodbAPI, waitRegion, instanceID, snapshot.ID, defaultMongodbSnapshotTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for MongoDB snapshot completion",
				fmt.Sprintf("Snapshot %s for instance %s did not reach a terminal state: %s", snapshot.ID, instanceID, err),
			)

			return
		}
	}
}
