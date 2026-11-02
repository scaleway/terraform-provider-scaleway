package mailbox

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	providertypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*AliasResource)(nil)
	_ resource.ResourceWithConfigure   = (*AliasResource)(nil)
	_ resource.ResourceWithImportState = (*AliasResource)(nil)
	_ resource.ResourceWithIdentity    = (*AliasResource)(nil)
)

func NewAliasResource() resource.Resource {
	return &AliasResource{}
}

type AliasResource struct {
	api  *mailboxsdk.API
	meta *meta.Meta
}

type aliasResourceModel struct {
	MailboxID   types.String `tfsdk:"mailbox_id"`
	LocalPart   types.String `tfsdk:"local_part"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type aliasResourceIdentityModel = framework.GlobalIdentity

func (r *AliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailbox_alias"
}

//go:embed descriptions/alias_resource.md
var aliasResourceDescription string

func (r *AliasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: aliasResourceDescription,
		Description:         aliasResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the mailbox alias",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mailbox_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the mailbox to which this alias belongs",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"local_part": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Local part of the alias email address (the part before the @). Changing this forces a new alias.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the alias. Omit the attribute to leave it empty; an empty string is not allowed.",
				Validators: []validator.String{
					// Reject "" so config never differs from API empty → null in state (perpetual diff).
					stringvalidator.LengthAtLeast(1),
				},
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Full email address of the alias (local_part@domain)",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Alias status (provisioning, ready, deleting, …)",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of alias creation (RFC 3339 format)",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of last update (RFC 3339 format)",
			},
		},
	}
}

func (r *AliasResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultGlobal()
}

func (r *AliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = mailboxsdk.NewAPI(r.meta.ScwClient())
}

func (r *AliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aliasResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &mailboxsdk.CreateAliasRequest{
		MailboxID: plan.MailboxID.ValueString(),
		LocalPart: plan.LocalPart.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && plan.Description.ValueString() != "" {
		createReq.Description = providertypes.ExpandStringPtr(plan.Description.ValueString())
	}

	alias, err := r.api.CreateAlias(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create mailbox alias", err.Error())

		return
	}

	alias, err = waitForAlias(ctx, r.api, alias.ID, defaultMailboxTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for mailbox alias", err.Error())

		return
	}

	state := convertAliasToState(alias)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(alias.ID))...)
}

func (r *AliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    aliasResourceModel
		identity aliasResourceIdentityModel
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

	aliasID := state.ID.ValueString()
	if identityAvailable {
		aliasID = identity.ID.ValueString()
	}

	if aliasID == "" {
		resp.Diagnostics.AddError("Alias ID not set", "Cannot refresh mailbox alias without an ID")

		return
	}

	alias, err := r.api.GetAlias(&mailboxsdk.GetAliasRequest{AliasID: aliasID}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to get mailbox alias", err.Error())

		return
	}

	state = convertAliasToState(alias)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(alias.ID))...)
}

func (r *AliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  aliasResourceModel
		state aliasResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		_, err := r.api.UpdateAlias(&mailboxsdk.UpdateAliasRequest{
			AliasID:     plan.ID.ValueString(),
			Description: providertypes.ExpandUpdatedStringPtr(plan.Description.ValueString()),
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update mailbox alias", err.Error())

			return
		}
	}

	alias, err := r.api.GetAlias(&mailboxsdk.GetAliasRequest{AliasID: plan.ID.ValueString()}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get mailbox alias after update", err.Error())

		return
	}

	newState := convertAliasToState(alias)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(alias.ID))...)
}

func (r *AliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aliasResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.api.DeleteAlias(&mailboxsdk.DeleteAliasRequest{AliasID: state.ID.ValueString()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete mailbox alias", err.Error())

		return
	}

	err = waitForAliasDeleted(ctx, r.api, state.ID.ValueString(), defaultMailboxTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for mailbox alias deletion", err.Error())
	}
}

func (r *AliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func convertAliasToState(alias *mailboxsdk.Alias) aliasResourceModel {
	description := types.StringNull()
	if alias.Description != "" {
		description = types.StringValue(alias.Description)
	}

	return aliasResourceModel{
		ID:          types.StringValue(alias.ID),
		MailboxID:   types.StringValue(alias.MailboxID),
		LocalPart:   types.StringValue(localPartFromEmail(alias.Email)),
		Description: description,
		Email:       types.StringValue(alias.Email),
		Status:      types.StringValue(alias.Status.String()),
		CreatedAt:   flattenTime(alias.CreatedAt),
		UpdatedAt:   flattenTime(alias.UpdatedAt),
	}
}
