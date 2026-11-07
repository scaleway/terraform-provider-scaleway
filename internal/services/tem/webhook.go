package tem

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/tem_webhook_resource.md
var webhookResourceDescription string

var (
	_ resource.Resource                = (*WebhookResource)(nil)
	_ resource.ResourceWithConfigure   = (*WebhookResource)(nil)
	_ resource.ResourceWithImportState = (*WebhookResource)(nil)
	_ resource.ResourceWithIdentity    = (*WebhookResource)(nil)
)

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

type WebhookResource struct {
	api  *tem.API
	meta *meta.Meta
}

type webhookResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DomainID       types.String `tfsdk:"domain_id"`
	Name           types.String `tfsdk:"name"`
	EventTypes     types.List   `tfsdk:"event_types"`
	SnsArn         types.String `tfsdk:"sns_arn"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	Region         types.String `tfsdk:"region"`
	ProjectID      types.String `tfsdk:"project_id"`
}

type webhookResourceIdentityModel = identityfw.RegionalIdentity

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tem_webhook"
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: webhookResourceDescription,
		Description:         webhookResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain id",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the webhook",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 127),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"event_types": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of event types",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.NoneOf("unknown_type")),
				},
			},
			"sns_arn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "SNS ARN",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 127),
				},
			},
			"organization_id": account.OrganizationIDSchemaAttribute(),
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp",
			},
			"region":     regional.SchemaAttributeComputed(),
			"project_id": account.ProjectIDSchemaAttribute(),
		},
	}
}

func (r *WebhookResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *WebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = tem.NewAPI(r.meta.ScwClient())
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
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

	eventTypes := expandWebhookEventTypes(ctx, plan.EventTypes, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.api.CreateWebhook(&tem.CreateWebhookRequest{
		Region:     region,
		ProjectID:  projectID,
		Name:       plan.Name.ValueString(),
		DomainID:   locality.ExpandID(plan.DomainID.ValueString()),
		SnsArn:     plan.SnsArn.ValueString(),
		EventTypes: eventTypes,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create TEM webhook", err.Error())

		return
	}

	// Mirror SDKv2 Create→Read so VCR cassette GET count stays aligned.
	webhook, err = r.api.GetWebhook(&tem.GetWebhookRequest{
		Region:    region,
		WebhookID: webhook.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read TEM webhook after create", err.Error())

		return
	}

	state := flattenWebhook(ctx, webhook, region, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, webhook.ID))...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    webhookResourceModel
		identity webhookResourceIdentityModel
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
			resp.Diagnostics.AddError("Failed to parse TEM webhook ID", err.Error())

			return
		}
	}

	webhook, err := r.api.GetWebhook(&tem.GetWebhookRequest{
		Region:    region,
		WebhookID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read TEM webhook", err.Error())

		return
	}

	newState := flattenWebhook(ctx, webhook, region, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, webhook.ID))...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  webhookResourceModel
		state webhookResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse TEM webhook ID", err.Error())

		return
	}

	updateReq := &tem.UpdateWebhookRequest{
		Region:    region,
		WebhookID: id,
	}
	hasChanged := false

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
		hasChanged = true
	}

	if !plan.EventTypes.Equal(state.EventTypes) {
		updateReq.EventTypes = expandWebhookEventTypes(ctx, plan.EventTypes, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		hasChanged = true
	}

	if !plan.SnsArn.Equal(state.SnsArn) {
		snsArn := plan.SnsArn.ValueString()
		updateReq.SnsArn = &snsArn
		hasChanged = true
	}

	if hasChanged {
		_, err = r.api.UpdateWebhook(updateReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update TEM webhook", err.Error())

			return
		}
	}

	webhook, err := r.api.GetWebhook(&tem.GetWebhookRequest{
		Region:    region,
		WebhookID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read TEM webhook after update", err.Error())

		return
	}

	newState := flattenWebhook(ctx, webhook, region, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, webhook.ID))...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse TEM webhook ID", err.Error())

		return
	}

	err = r.api.DeleteWebhook(&tem.DeleteWebhookRequest{
		WebhookID: id,
		Region:    region,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete TEM webhook", err.Error())
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func flattenWebhook(ctx context.Context, webhook *tem.Webhook, region scw.Region, diags *diag.Diagnostics) webhookResourceModel {
	eventTypes := make([]string, len(webhook.EventTypes))
	for i, v := range webhook.EventTypes {
		eventTypes[i] = string(v)
	}

	eventTypesList, d := types.ListValueFrom(ctx, types.StringType, eventTypes)
	diags.Append(d...)

	return webhookResourceModel{
		ID:             types.StringValue(regional.NewIDString(region, webhook.ID)),
		Name:           types.StringValue(webhook.Name),
		DomainID:       types.StringValue(regional.NewIDString(region, webhook.DomainID)),
		OrganizationID: types.StringValue(webhook.OrganizationID),
		ProjectID:      types.StringValue(webhook.ProjectID),
		EventTypes:     eventTypesList,
		SnsArn:         types.StringValue(webhook.SnsArn),
		CreatedAt:      types.StringValue(scwtypes.FlattenTime(webhook.CreatedAt).(string)),
		UpdatedAt:      types.StringValue(scwtypes.FlattenTime(webhook.UpdatedAt).(string)),
		Region:         types.StringValue(region.String()),
	}
}
