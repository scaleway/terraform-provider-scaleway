package cockpit

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_exporter_resource.md
var cockpitExporterResourceDescription string

var (
	_ resource.Resource                = (*ExporterResource)(nil)
	_ resource.ResourceWithConfigure   = (*ExporterResource)(nil)
	_ resource.ResourceWithImportState = (*ExporterResource)(nil)
	_ resource.ResourceWithIdentity    = (*ExporterResource)(nil)
)

func NewExporterResource() resource.Resource {
	return &ExporterResource{}
}

type ExporterResource struct {
	api  *cockpit.RegionalAPI
	meta *meta.Meta
}

type exporterDatadogModel struct {
	APIKey   types.String `tfsdk:"api_key"`
	Endpoint types.String `tfsdk:"endpoint"`
}

type exporterOTLPModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Headers  types.Map    `tfsdk:"headers"`
}

type exporterResourceModel struct {
	Timeouts           timeouts.Value         `tfsdk:"timeouts"`
	ExportedProducts   types.List             `tfsdk:"exported_products"`
	ID                 types.String           `tfsdk:"id"`
	Name               types.String           `tfsdk:"name"`
	Description        types.String           `tfsdk:"description"`
	DatasourceID       types.String           `tfsdk:"datasource_id"`
	Status             types.String           `tfsdk:"status"`
	CreatedAt          types.String           `tfsdk:"created_at"`
	UpdatedAt          types.String           `tfsdk:"updated_at"`
	SRN                types.String           `tfsdk:"srn"`
	ProjectID          types.String           `tfsdk:"project_id"`
	Region             types.String           `tfsdk:"region"`
	DatadogDestination []exporterDatadogModel `tfsdk:"datadog_destination"`
	OtlpDestination    []exporterOTLPModel    `tfsdk:"otlp_destination"`
}

func (r *ExporterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_exporter"
}

// suppressLocalityDiff keeps the prior value when the configured ID only differs by its locality prefix.
type suppressLocalityDiff struct{}

func (suppressLocalityDiff) Description(_ context.Context) string {
	return "Suppresses diffs between a bare UUID and a regional ID."
}

func (m suppressLocalityDiff) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (suppressLocalityDiff) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	if locality.ExpandID(req.StateValue.ValueString()) == locality.ExpandID(req.PlanValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// exportedProductsAll is the default value of exported_products.
func exportedProductsAll() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{types.StringValue("all")})
}

func (r *ExporterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitExporterResourceDescription,
		Description:         cockpitExporterResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the data export",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the data export",
			},
			"datasource_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the data source linked to the data export",
				PlanModifiers: []planmodifier.String{
					suppressLocalityDiff{},
				},
			},
			"exported_products": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(exportedProductsAll()),
				MarkdownDescription: "List of Scaleway products to export. Use [\"all\"] to export all products. Use scaleway_cockpit_products data source for valid product names.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the data export",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of the creation of the data export (RFC 3339 format)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of the last update of the data export (RFC 3339 format)",
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the data export",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": account.ProjectIDSchemaFw("The project_id you want to attach the resource to"),
			"region":     regional.SchemaAttributeComputed(),
		},
		Blocks: map[string]schema.Block{
			"datadog_destination": schema.ListNestedBlock{
				MarkdownDescription: "Datadog destination configuration",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"api_key": schema.StringAttribute{
							Required:            true,
							Sensitive:           true,
							MarkdownDescription: "Datadog API key",
						},
						"endpoint": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Datadog endpoint URL",
						},
					},
				},
			},
			"otlp_destination": schema.ListNestedBlock{
				MarkdownDescription: "OTLP destination configuration",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"endpoint": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "OTLP endpoint URL",
						},
						"headers": schema.MapAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Headers to include in requests",
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *ExporterResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *ExporterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// expandExporterDestinations returns the Datadog or OTLP destination of the plan; exactly one must be set.
func expandExporterDestinations(ctx context.Context, plan exporterResourceModel) (*cockpit.ExporterDatadogDestination, *cockpit.ExporterOTLPDestination, error) {
	hasDatadog := len(plan.DatadogDestination) > 0
	hasOTLP := len(plan.OtlpDestination) > 0

	switch {
	case hasDatadog && hasOTLP:
		return nil, nil, errors.New("cannot specify both datadog_destination and otlp_destination")
	case !hasDatadog && !hasOTLP:
		return nil, nil, errors.New("must specify either datadog_destination or otlp_destination")
	case hasDatadog:
		dest := &cockpit.ExporterDatadogDestination{APIKey: plan.DatadogDestination[0].APIKey.ValueStringPointer()}

		if endpoint := plan.DatadogDestination[0].Endpoint.ValueString(); endpoint != "" {
			dest.Endpoint = &endpoint
		}

		return dest, nil, nil
	}

	dest := &cockpit.ExporterOTLPDestination{
		Endpoint: plan.OtlpDestination[0].Endpoint.ValueString(),
		Headers:  map[string]string{},
	}

	if !plan.OtlpDestination[0].Headers.IsNull() {
		diags := plan.OtlpDestination[0].Headers.ElementsAs(ctx, &dest.Headers, false)
		if diags.HasError() {
			return nil, nil, errors.New(diags[0].Detail())
		}
	}

	return nil, dest, nil
}

func expandExportedProducts(ctx context.Context, products types.List) ([]string, diag.Diagnostics) {
	var result []string

	diags := products.ElementsAs(ctx, &result, false)

	return result, diags
}

func (r *ExporterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan exporterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	products, diags := expandExportedProducts(ctx, plan.ExportedProducts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	datadog, otlp, err := expandExporterDestinations(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid exporter destination", err.Error())

		return
	}

	createReq := &cockpit.RegionalAPICreateExporterRequest{
		Region:             region,
		DatasourceID:       locality.ExpandID(plan.DatasourceID.ValueString()),
		Name:               plan.Name.ValueString(),
		ExportedProducts:   products,
		DatadogDestination: datadog,
		OtlpDestination:    otlp,
	}

	if plan.Description.ValueString() != "" {
		createReq.Description = plan.Description.ValueStringPointer()
	}

	res, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Exporter, error) {
		return r.api.CreateExporter(createReq, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Cockpit exporter", err.Error())

		return
	}

	exporter, err := waitForExporter(ctx, r.api, region, res.ID, timeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for Cockpit exporter", err.Error())

		return
	}

	r.refresh(ctx, region, exporter.ID, plan, &resp.State, resp.Identity, &resp.Diagnostics)
}

// exporterProjectID returns the project of the data source linked to the exporter.
func exporterProjectID(ctx context.Context, api *cockpit.RegionalAPI, res *cockpit.Exporter) (string, error) {
	ds, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.DataSource, error) {
		return api.GetDataSource(&cockpit.RegionalAPIGetDataSourceRequest{
			Region:       res.Region,
			DataSourceID: res.DatasourceID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		return "", err
	}

	return ds.ProjectID, nil
}

// optionalString maps an empty API value to null when the optional attribute was not set.
func optionalString(value string, prior types.String) types.String {
	if value == "" && prior.IsNull() {
		return types.StringNull()
	}

	return types.StringValue(value)
}

// flattenExporterDestinations builds the destination blocks; priors provide the Datadog API key
// (never returned by the API) and tell whether optional values were configured.
func flattenExporterDestinations(ctx context.Context, res *cockpit.Exporter, priorDatadog []exporterDatadogModel, priorOTLP []exporterOTLPModel) ([]exporterDatadogModel, []exporterOTLPModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	datadog := []exporterDatadogModel{}
	otlp := []exporterOTLPModel{}

	if res.DatadogDestination != nil {
		prior := exporterDatadogModel{APIKey: types.StringNull(), Endpoint: types.StringNull()}
		if len(priorDatadog) > 0 {
			prior = priorDatadog[0]
		}

		endpoint := ""
		if res.DatadogDestination.Endpoint != nil {
			endpoint = *res.DatadogDestination.Endpoint
		}

		datadog = append(datadog, exporterDatadogModel{
			APIKey:   prior.APIKey,
			Endpoint: optionalString(endpoint, prior.Endpoint),
		})
	}

	if res.OtlpDestination != nil {
		headers := types.MapNull(types.StringType)

		if len(res.OtlpDestination.Headers) > 0 || (len(priorOTLP) > 0 && !priorOTLP[0].Headers.IsNull()) {
			var d diag.Diagnostics

			headers, d = types.MapValueFrom(ctx, types.StringType, res.OtlpDestination.Headers)
			diags.Append(d...)
		}

		otlp = append(otlp, exporterOTLPModel{
			Endpoint: types.StringValue(res.OtlpDestination.Endpoint),
			Headers:  headers,
		})
	}

	return datadog, otlp, diags
}

func (r *ExporterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	identity, identityAvailable := getRegionalIdentity(ctx, req.Identity)

	var state exporterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := resolveRegionalID(identity, identityAvailable, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Cockpit exporter ID", err.Error())

		return
	}

	timeout, diags := state.Timeouts.Read(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	readCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := r.getExporter(readCtx, region, id)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Cockpit exporter", err.Error())

		return
	}

	newState, diags := r.flattenExporter(readCtx, res, state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(res.Region, res.ID))...)
}

func (r *ExporterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state exporterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Cockpit exporter ID", err.Error())

		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	datadog, otlp, err := expandExporterDestinations(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid exporter destination", err.Error())

		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &cockpit.RegionalAPIUpdateExporterRequest{
		Region:             region,
		ExporterID:         id,
		DatadogDestination: datadog,
		OtlpDestination:    otlp,
	}

	if !plan.Name.Equal(state.Name) {
		updateReq.Name = plan.Name.ValueStringPointer()
	}

	if !plan.Description.Equal(state.Description) {
		updateReq.Description = plan.Description.ValueStringPointer()
	}

	if !plan.ExportedProducts.Equal(state.ExportedProducts) {
		products, d := expandExportedProducts(ctx, plan.ExportedProducts)
		resp.Diagnostics.Append(d...)

		updateReq.ExportedProducts = &products
	}

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = waitForExporter(ctx, r.api, region, id, timeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for Cockpit exporter", err.Error())

		return
	}

	_, err = transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Exporter, error) {
		return r.api.UpdateExporter(updateReq, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Cockpit exporter", err.Error())

		return
	}

	_, err = waitForExporter(ctx, r.api, region, id, timeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for Cockpit exporter", err.Error())

		return
	}

	r.refresh(ctx, region, id, plan, &resp.State, resp.Identity, &resp.Diagnostics)
}

func (r *ExporterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	identity, identityAvailable := getRegionalIdentity(ctx, req.Identity)

	var state exporterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := resolveRegionalID(identity, identityAvailable, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Cockpit exporter ID", err.Error())

		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = waitForExporter(ctx, r.api, region, id, timeout)
	if httperrors.Is404(err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for Cockpit exporter", err.Error())

		return
	}

	err = transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		return r.api.DeleteExporter(&cockpit.RegionalAPIDeleteExporterRequest{
			Region:     region,
			ExporterID: id,
		}, scw.WithContext(ctx))
	})
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete Cockpit exporter", err.Error())
	}
}

func (r *ExporterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

// refresh reads the exporter and stores it, with its identity, in state; prior carries the planned values.
func (r *ExporterResource) refresh(ctx context.Context, region scw.Region, id string, prior exporterResourceModel, state *tfsdk.State, identity *tfsdk.ResourceIdentity, diags *diag.Diagnostics) {
	res, err := r.getExporter(ctx, region, id)
	if err != nil {
		diags.AddError("Failed to read Cockpit exporter", err.Error())

		return
	}

	model, d := r.flattenExporter(ctx, res, prior)
	diags.Append(d...)

	if diags.HasError() {
		return
	}

	diags.Append(state.Set(ctx, &model)...)
	diags.Append(identity.Set(ctx, identityfw.SetRegionalIdentity(res.Region, res.ID))...)
}

// getExporter returns the exporter; the API response may omit the region, so it is set from the request.
func (r *ExporterResource) getExporter(ctx context.Context, region scw.Region, id string) (*cockpit.Exporter, error) {
	res, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Exporter, error) {
		return r.api.GetExporter(&cockpit.RegionalAPIGetExporterRequest{
			Region:     region,
			ExporterID: id,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		return nil, err
	}

	res.Region = region

	return res, nil
}

// flattenExporter builds the model from the API; prior carries values the API does not return (api_key)
// or may return differently from the configuration (locality of datasource_id, empty optional values).
func (r *ExporterResource) flattenExporter(ctx context.Context, res *cockpit.Exporter, prior exporterResourceModel) (exporterResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	datasourceID := types.StringValue(regional.NewIDString(res.Region, res.DatasourceID))
	if !prior.DatasourceID.IsNull() && locality.ExpandID(prior.DatasourceID.ValueString()) == res.DatasourceID {
		datasourceID = prior.DatasourceID
	}

	products, d := types.ListValueFrom(ctx, types.StringType, res.ExportedProducts)
	diags.Append(d...)

	datadog, otlp, d := flattenExporterDestinations(ctx, res, prior.DatadogDestination, prior.OtlpDestination)
	diags.Append(d...)

	model := exporterResourceModel{
		Timeouts:           prior.Timeouts,
		DatadogDestination: datadog,
		OtlpDestination:    otlp,
		ExportedProducts:   products,
		ID:                 types.StringValue(regional.NewIDString(res.Region, res.ID)),
		Name:               types.StringValue(res.Name),
		Description:        optionalString(res.Description, prior.Description),
		DatasourceID:       datasourceID,
		Status:             types.StringValue(res.Status.String()),
		CreatedAt:          flattenCockpitTime(res.CreatedAt),
		UpdatedAt:          flattenCockpitTime(res.UpdatedAt),
		SRN:                types.StringValue(res.Srn),
		Region:             types.StringValue(res.Region.String()),
		ProjectID:          prior.ProjectID,
	}

	projectID, err := exporterProjectID(ctx, r.api, res)
	switch {
	case err == nil:
		model.ProjectID = types.StringValue(projectID)
	case model.ProjectID.IsUnknown():
		model.ProjectID = types.StringNull()
	}

	return model, diags
}
