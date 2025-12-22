package instance

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
)

var (
	_ action.Action              = (*CreateSnapshot)(nil)
	_ action.ActionWithConfigure = (*CreateSnapshot)(nil)
)

type CreateSnapshot struct {
	blockAndInstanceAPI *instancehelpers.BlockAndInstanceAPI
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
	c.blockAndInstanceAPI = instancehelpers.NewBlockAndInstanceAPI(client)
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

//go:embed descriptions/create_snapshot_action.md
var createSnapshotActionDescription string

func (c *CreateSnapshot) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: createSnapshotActionDescription,
		Description:         createSnapshotActionDescription,
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
				Description: "Wait for snapshotting operation to be completed",
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

	if c.blockAndInstanceAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured instanceAPI / blockAPI",
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

	volume, err := c.blockAndInstanceAPI.GetUnknownVolume(&instancehelpers.GetUnknownVolumeRequest{
		VolumeID: volumeID,
		Zone:     scw.Zone(zone),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"could not find volume "+data.VolumeID.ValueString(),
			err.Error(),
		)

		return
	}

	switch volume.InstanceVolumeType {
	case instance.VolumeVolumeTypeLSSD:
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

		snapshot, err := c.blockAndInstanceAPI.CreateSnapshot(actionReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"error creating instance snapshot",
				err.Error())

			return
		}

		if data.Wait.ValueBool() {
			_, errWait := c.blockAndInstanceAPI.WaitForSnapshot(&instance.WaitForSnapshotRequest{
				SnapshotID: snapshot.Snapshot.ID,
				Zone:       scw.Zone(zone),
			}, scw.WithContext(ctx))
			if errWait != nil {
				resp.Diagnostics.AddError(
					"error waiting for instance snapshot",
					err.Error())
			}
		}
	case instance.VolumeVolumeTypeSbsVolume:
		api := c.blockAndInstanceAPI.BlockAPI

		actionReq := &block.CreateSnapshotRequest{
			VolumeID: volumeID,
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
				actionReq.Tags = tags
			}
		}

		snapshot, err := api.CreateSnapshot(actionReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"error creating block snapshot",
				err.Error())

			return
		}

		if data.Wait.ValueBool() {
			_, errWait := api.WaitForSnapshot(&block.WaitForSnapshotRequest{
				SnapshotID: snapshot.ID,
				Zone:       scw.Zone(zone),
			}, scw.WithContext(ctx))
			if errWait != nil {
				resp.Diagnostics.AddError(
					"error waiting for block snapshot",
					err.Error())
			}
		}
	case instance.VolumeVolumeTypeScratch:
		resp.Diagnostics.AddError(
			"invalid volume type",
			"cannot create snapshot from a volume of type scratch",
		)
	default:
		resp.Diagnostics.AddError(
			"invalid volume type",
			fmt.Sprintf("unknown volume type %q", volume.InstanceVolumeType),
		)
	}
}
