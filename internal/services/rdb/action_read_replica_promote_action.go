package rdb

import (
	"context"
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
	_ action.Action              = (*ReadReplicaPromoteAction)(nil)
	_ action.ActionWithConfigure = (*ReadReplicaPromoteAction)(nil)
)

// ReadReplicaPromoteAction promotes a read replica to a standalone instance.
type ReadReplicaPromoteAction struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

func (a *ReadReplicaPromoteAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *ReadReplicaPromoteAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_read_replica_promote_action"
}

type ReadReplicaPromoteActionModel struct {
	ReadReplicaID types.String `tfsdk:"read_replica_id"`
	Region        types.String `tfsdk:"region"`
	Wait          types.Bool   `tfsdk:"wait"`
}

// NewReadReplicaPromoteAction returns a new RDB read replica promote action.
func NewReadReplicaPromoteAction() action.Action {
	return &ReadReplicaPromoteAction{}
}

func (a *ReadReplicaPromoteAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"read_replica_id": schema.StringAttribute{
				Required:    true,
				Description: "RDB read replica ID to promote. Can be a plain UUID or a regional ID.",
			},
			"region": regional.SchemaAttribute(),
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the read replica promotion to complete before returning.",
			},
		},
	}
}

func (a *ReadReplicaPromoteAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ReadReplicaPromoteActionModel

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

	if data.ReadReplicaID.IsNull() || data.ReadReplicaID.IsUnknown() || data.ReadReplicaID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing read_replica_id",
			"The read_replica_id attribute is required to promote a read replica.",
		)

		return
	}

	readReplicaID := locality.ExpandID(data.ReadReplicaID.ValueString())

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
		if derivedRegion, id, parseErr := regional.ParseID(data.ReadReplicaID.ValueString()); parseErr == nil {
			region = derivedRegion
			readReplicaID = id
		} else if a.meta != nil {
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Unable to determine region",
					"Failed to get default region from provider configuration. Please set the region attribute, use a regional read_replica_id, or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	if region == "" {
		resp.Diagnostics.AddError(
			"Missing region",
			"Could not determine region for RDB read replica promotion. Please set the region attribute, use a regional read_replica_id, or configure a default region in the provider.",
		)

		return
	}

	promoteReq := &rdb.PromoteReadReplicaRequest{
		Region:        region,
		ReadReplicaID: readReplicaID,
	}

	promotedInstance, err := a.rdbAPI.PromoteReadReplica(promoteReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing RDB PromoteReadReplica action",
			fmt.Sprintf("Failed to promote read replica %s: %s", readReplicaID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		newInstanceID := promotedInstance.ID

		_, err = waitForRDBInstance(ctx, a.rdbAPI, region, newInstanceID, defaultInstanceTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for RDB read replica promotion completion",
				fmt.Sprintf("Read replica %s promotion did not complete: %s", readReplicaID, err),
			)

			return
		}
	}
}
