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
	_ action.Action              = (*InstanceLogPrepareAction)(nil)
	_ action.ActionWithConfigure = (*InstanceLogPrepareAction)(nil)
)

// InstanceLogPrepareAction prepares a log for an RDB instance.
type InstanceLogPrepareAction struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

func (a *InstanceLogPrepareAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *InstanceLogPrepareAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_instance_prepare_logs"
}

type InstanceLogPrepareActionModel struct {
	InstanceID types.String `tfsdk:"instance_id"`
	Region     types.String `tfsdk:"region"`
	StartDate  types.String `tfsdk:"start_date"`
	EndDate    types.String `tfsdk:"end_date"`
	Wait       types.Bool   `tfsdk:"wait"`
}

// NewInstanceLogPrepareAction returns a new RDB instance log prepare action.
func NewInstanceLogPrepareAction() action.Action {
	return &InstanceLogPrepareAction{}
}

//go:embed descriptions/instance_log_prepare_action.md
var instanceLogPrepareDescription string

func (a *InstanceLogPrepareAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: instanceLogPrepareDescription,
		Description:         instanceLogPrepareDescription,
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "RDB instance ID to prepare log for. Can be a plain UUID or a regional ID.",
			},
			"region": regional.SchemaAttribute(),
			"start_date": schema.StringAttribute{
				Optional:    true,
				Description: "Start datetime of the log in RFC3339 format (ISO 8601).",
			},
			"end_date": schema.StringAttribute{
				Optional:    true,
				Description: "End datetime of the log in RFC3339 format (ISO 8601).",
			},
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the log preparation to complete before returning.",
			},
		},
	}
}

func (a *InstanceLogPrepareAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data InstanceLogPrepareActionModel

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
			"The instance_id attribute is required to prepare an RDB instance log.",
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
			"Could not determine region for RDB instance log preparation. Please set the region attribute, use a regional instance_id, or configure a default region in the provider.",
		)

		return
	}

	prepareReq := &rdb.PrepareInstanceLogsRequest{
		Region:     region,
		InstanceID: instanceID,
	}

	if !data.StartDate.IsNull() && !data.StartDate.IsUnknown() && data.StartDate.ValueString() != "" {
		parsedStartDate, err := time.Parse(time.RFC3339, data.StartDate.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid start_date value",
				fmt.Sprintf("The start_date attribute must be a valid RFC3339 timestamp. Got %q: %s", data.StartDate.ValueString(), err),
			)

			return
		}

		prepareReq.StartDate = &parsedStartDate
	}

	if !data.EndDate.IsNull() && !data.EndDate.IsUnknown() && data.EndDate.ValueString() != "" {
		parsedEndDate, err := time.Parse(time.RFC3339, data.EndDate.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid end_date value",
				fmt.Sprintf("The end_date attribute must be a valid RFC3339 timestamp. Got %q: %s", data.EndDate.ValueString(), err),
			)

			return
		}

		prepareReq.EndDate = &parsedEndDate
	}

	_, err := a.rdbAPI.PrepareInstanceLogs(prepareReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing RDB PrepareInstanceLogs action",
			fmt.Sprintf("Failed to prepare logs for instance %s: %s", instanceID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		_, err = waitForRDBInstance(ctx, a.rdbAPI, region, instanceID, defaultInstanceTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for RDB log preparation completion",
				fmt.Sprintf("Log preparation for instance %s did not complete: %s", instanceID, err),
			)

			return
		}
	}
}
