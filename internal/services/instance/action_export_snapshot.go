package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*ExportSnapshot)(nil)
	_ action.ActionWithConfigure = (*ExportSnapshot)(nil)
)

type ExportSnapshot struct {
	instanceAPI *instance.API
}

func (e *ExportSnapshot) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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
	e.instanceAPI = instance.NewAPI(client)
}

func (e *ExportSnapshot) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_export_snapshot"
}

type ExportSnapshotModel struct {
	Zone       types.String `tfsdk:"zone"`
	SnapshotID types.String `tfsdk:"snapshot_id"`
	Bucket     types.String `tfsdk:"bucket"`
	Key        types.String `tfsdk:"key"`
	Wait       types.Bool   `tfsdk:"wait"`
}

func NewExportSnapshot() action.Action {
	return &ExportSnapshot{}
}

func (e *ExportSnapshot) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"snapshot_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the snapshot to export",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "Zone of the snapshot to export",
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket to export the snapshot to",
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "Object key to save the snapshot to",
			},
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for exporting operation to be completed",
			},
		},
	}
}

func (e *ExportSnapshot) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ExportSnapshotModel
	// Read action config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if e.instanceAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured instanceAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	zone, snapshotID, _ := locality.ParseLocalizedID(data.SnapshotID.ValueString())
	if zone == "" {
		if !data.Zone.IsNull() {
			zone = data.Zone.ValueString()
		} else {
			resp.Diagnostics.AddError(
				"missing zone in config",
				fmt.Sprintf("zone could not be extracted from either the action configuration or the resource ID (%s)",
					data.SnapshotID.ValueString(),
				),
			)

			return
		}
	}

	actionReq := &instance.ExportSnapshotRequest{
		SnapshotID: snapshotID,
		Zone:       scw.Zone(zone),
		Bucket:     data.Bucket.ValueString(),
		Key:        data.Key.ValueString(),
	}
	if !data.Zone.IsNull() {
		actionReq.Zone = scw.Zone(data.Zone.ValueString())
	}

	_, err := e.instanceAPI.ExportSnapshot(actionReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"error exporting snapshot",
			fmt.Sprintf("%s", err))

		return
	}

	if data.Wait.ValueBool() {
		_, err = e.instanceAPI.WaitForSnapshot(&instance.WaitForSnapshotRequest{
			SnapshotID: snapshotID,
			Zone:       scw.Zone(zone),
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"error waiting snapshot",
				fmt.Sprintf("%s", err))
		}
	}
}
