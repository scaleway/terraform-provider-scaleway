package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*CreateSnapshot)(nil)
	_ action.ActionWithConfigure = (*CreateSnapshot)(nil)
)

type CreateSnapshot struct {
	instanceAPI *instance.API
}

func (c *CreateSnapshot) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *scw.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client := m.ScwClient()
	c.instanceAPI = instance.NewAPI(client)
}

func (c *CreateSnapshot) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_create_snapshot"
}

type CreateSnapshotModel struct {
	Zone     types.String `tfsdk:"zone"`
	VolumeID types.String `tfsdk:"volume_id"`
	Name     types.String `tfsdk:"name"`
	Tags     types.List   `tfsdk:"tags"`
	Wait     types.Bool   `tfsdk:"wait"`
}

func NewCreateSnapshot() action.Action {
	return &CreateSnapshot{}
}

func (c *CreateSnapshot) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"volume_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the volume to snapshot",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "Zone of the volume to snapshot",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the snapshot",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: basetypes.StringType{},
				Description: "List of tags associated with the snapshot",
			},
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for snapshotting operation to be finished",
			},
		},
	}
}

func (c *CreateSnapshot) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data CreateSnapshotModel
	// Read action config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if c.instanceAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured instanceAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	zone, volumeID, _ := locality.ParseLocalizedID(data.VolumeID.ValueString())
	if zone == "" {
		if !data.Zone.IsNull() {
			zone = data.Zone.ValueString()
		} else {
			resp.Diagnostics.AddError(
				"missing zone in config",
				fmt.Sprintf("zone could not be extracted from either the action configuration or the resource ID (%s)",
					data.VolumeID.ValueString(),
				),
			)

			return
		}
	}

	actionReq := &instance.CreateSnapshotRequest{
		VolumeID: &volumeID,
		Zone:     scw.Zone(zone),
	}

	if !data.Name.IsNull() {
		actionReq.Name = data.Name.ValueString()
	}

	if len(data.Tags.Elements()) > 0 {
		tags := make([]string, 0, len(data.Tags.Elements()))

		diags := data.Tags.ElementsAs(ctx, &tags, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
		} else {
			actionReq.Tags = &tags
		}
	}

	snapshot, err := c.instanceAPI.CreateSnapshot(actionReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating snapshot",
			fmt.Sprintf("%s", err))

		return
	}

	if data.Wait.ValueBool() {
		_, errWait := c.instanceAPI.WaitForSnapshot(&instance.WaitForSnapshotRequest{
			SnapshotID: snapshot.Snapshot.ID,
			Zone:       scw.Zone(zone),
		}, scw.WithContext(ctx))
		if errWait != nil {
			resp.Diagnostics.AddError(
				"error waiting for snapshot",
				fmt.Sprintf("%s", err))
		}
	}
}
