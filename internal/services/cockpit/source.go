package cockpit

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/cockpit_source_resource.md
var cockpitSourceResourceDescription string

var (
	_ resource.Resource                = (*SourceResource)(nil)
	_ resource.ResourceWithConfigure   = (*SourceResource)(nil)
	_ resource.ResourceWithImportState = (*SourceResource)(nil)
	_ resource.ResourceWithIdentity    = (*SourceResource)(nil)
)

func NewSourceResource() resource.Resource {
	return &SourceResource{}
}

type SourceResource struct {
	api  *cockpit.RegionalAPI
	meta *meta.Meta
}

type sourceResourceModel struct {
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
	ID                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	Type                    types.String   `tfsdk:"type"`
	URL                     types.String   `tfsdk:"url"`
	Origin                  types.String   `tfsdk:"origin"`
	Region                  types.String   `tfsdk:"region"`
	ProjectID               types.String   `tfsdk:"project_id"`
	CreatedAt               types.String   `tfsdk:"created_at"`
	UpdatedAt               types.String   `tfsdk:"updated_at"`
	PushURL                 types.String   `tfsdk:"push_url"`
	SRN                     types.String   `tfsdk:"srn"`
	RetentionDays           types.Int64    `tfsdk:"retention_days"`
	SynchronizedWithGrafana types.Bool     `tfsdk:"synchronized_with_grafana"`
}

type sourceResourceIdentityModel = identityfw.RegionalIdentity

func (r *SourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_source"
}

func (r *SourceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitSourceResourceDescription,
		Description:         cockpitSourceResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the datasource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The type of the datasource",
				Validators: []validator.String{
					verify.FrameworkValidateEnum[cockpit.DataSourceType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"retention_days": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The number of days to retain data. Use scaleway_cockpit_config data source to read allowed min, max, and default values for each data source type.",
				Validators: []validator.Int64{
					int64validator.Between(1, 1825),
				},
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the datasource",
			},
			"origin": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The origin of the datasource",
			},
			"synchronized_with_grafana": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates whether the data source is synchronized with Grafana",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time of the creation of the cockpit datasource",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time of the last update of the cockpit datasource",
			},
			"push_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL endpoint used for pushing data to the cockpit data source.",
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the data source",
			},
			"region": regional.SchemaAttributeComputed(),
			"project_id": account.ProjectIDSchemaFw(
				"The project_id you want to attach the resource to",
			),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *SourceResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *SourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
	r.api = cockpit.NewRegionalAPI(m.ScwClient())
}

func (r *SourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	createCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := transport.RetryOnStatusValue(createCtx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.DataSource, error) {
		return r.api.CreateDataSource(&cockpit.RegionalAPICreateDataSourceRequest{
			Region:        region,
			ProjectID:     projectID,
			Name:          plan.Name.ValueString(),
			Type:          cockpit.DataSourceType(plan.Type.ValueString()),
			RetentionDays: new(uint32(plan.RetentionDays.ValueInt64())),
		}, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Cockpit source", err.Error())

		return
	}

	state, diags := flattenSource(res, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(res.Region, res.ID))...)
}

func (r *SourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    sourceResourceModel
		identity sourceResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		id     string
		region scw.Region
		err    error
	)

	if identityAvailable {
		region, id = identityfw.ResolveRegionalIdentity(identity)
	} else {
		region, id, err = regional.ParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse Cockpit source ID", err.Error())

			return
		}
	}

	timeout, diags := state.Timeouts.Read(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	readCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := transport.RetryOnStatusValue(readCtx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.DataSource, error) {
		return r.api.GetDataSource(&cockpit.RegionalAPIGetDataSourceRequest{
			Region:       region,
			DataSourceID: id,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Cockpit source", err.Error())

		return
	}

	newState, diags := flattenSource(res, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(res.Region, res.ID))...)
}

func (r *SourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  sourceResourceModel
		state sourceResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Cockpit source ID", err.Error())

		return
	}

	updateRequest := &cockpit.RegionalAPIUpdateDataSourceRequest{
		DataSourceID: id,
		Region:       region,
	}
	hasChanges := false

	if !plan.Name.Equal(state.Name) {
		updateRequest.Name = new(plan.Name.ValueString())
		hasChanges = true
	}

	if !plan.RetentionDays.Equal(state.RetentionDays) {
		updateRequest.RetentionDays = new(uint32(plan.RetentionDays.ValueInt64()))
		hasChanges = true
	}

	var res *cockpit.DataSource

	if hasChanges {
		res, err = transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.DataSource, error) {
			return r.api.UpdateDataSource(updateRequest, scw.WithContext(ctx))
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to update Cockpit source", err.Error())

			return
		}
	} else {
		res, err = transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.DataSource, error) {
			return r.api.GetDataSource(&cockpit.RegionalAPIGetDataSourceRequest{
				Region:       region,
				DataSourceID: id,
			}, scw.WithContext(ctx))
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to read Cockpit source after update", err.Error())

			return
		}
	}

	newState, diags := flattenSource(res, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(res.Region, res.ID))...)
}

func (r *SourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		state    sourceResourceModel
		identity sourceResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		id     string
		region scw.Region
		err    error
	)

	if identityAvailable {
		region, id = identityfw.ResolveRegionalIdentity(identity)
	} else {
		region, id, err = regional.ParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse Cockpit source ID", err.Error())

			return
		}
	}

	timeout, diags := state.Timeouts.Delete(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = transport.RetryOnStatus(deleteCtx, transport.IAMPropagation, func(ctx context.Context) error {
		return r.api.DeleteDataSource(&cockpit.RegionalAPIDeleteDataSourceRequest{
			DataSourceID: id,
			Region:       region,
		}, scw.WithContext(ctx))
	})
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete Cockpit source", err.Error())
	}
}

func (r *SourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func flattenSource(res *cockpit.DataSource, timeoutsValue timeouts.Value) (sourceResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	pushURL, err := createCockpitPushURL(res.Type, res.URL)
	if err != nil {
		diags.AddError("Failed to build Cockpit source push URL", err.Error())

		return sourceResourceModel{}, diags
	}

	model := sourceResourceModel{
		Timeouts:                timeoutsValue,
		ID:                      types.StringValue(regional.NewIDString(res.Region, res.ID)),
		Name:                    types.StringValue(res.Name),
		Type:                    types.StringValue(res.Type.String()),
		URL:                     types.StringValue(res.URL),
		Origin:                  types.StringValue(res.Origin.String()),
		SynchronizedWithGrafana: types.BoolValue(res.SynchronizedWithGrafana),
		Region:                  types.StringValue(res.Region.String()),
		ProjectID:               types.StringValue(res.ProjectID),
		PushURL:                 types.StringValue(pushURL),
		SRN:                     types.StringValue(res.Srn),
		RetentionDays:           types.Int64Value(int64(res.RetentionDays)),
	}

	if res.CreatedAt != nil {
		model.CreatedAt = types.StringValue(res.CreatedAt.Format(time.RFC3339))
	}

	if res.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(res.UpdatedAt.Format(time.RFC3339))
	}

	return model, diags
}
