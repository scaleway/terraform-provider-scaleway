package mailbox

import (
	"context"
	_ "embed"
	"fmt"

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
	_ resource.Resource                = (*ForwardingResource)(nil)
	_ resource.ResourceWithConfigure   = (*ForwardingResource)(nil)
	_ resource.ResourceWithImportState = (*ForwardingResource)(nil)
	_ resource.ResourceWithIdentity    = (*ForwardingResource)(nil)
)

func NewForwardingResource() resource.Resource {
	return &ForwardingResource{}
}

type ForwardingResource struct {
	api  *mailboxsdk.API
	meta *meta.Meta
}

type forwardingResourceModel struct {
	MailboxID types.String `tfsdk:"mailbox_id"`
	Email     types.String `tfsdk:"email"`
	ID        types.String `tfsdk:"id"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type forwardingResourceIdentityModel = framework.GlobalIdentity

func (r *ForwardingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailbox_forwarding"
}

//go:embed descriptions/forwarding_resource.md
var forwardingResourceDescription string

func (r *ForwardingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: forwardingResourceDescription,
		Description:         forwardingResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the mailbox forwarding rule",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mailbox_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the mailbox for which to create the forwarding rule",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Destination email address. Must differ from the mailbox's own address. Can be updated in place.",
				Validators: []validator.String{
					verify.IsStringEmail(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Forwarding rule status (provisioning, ready, updating, deleting, …)",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of forwarding rule creation (RFC 3339 format)",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of last update (RFC 3339 format)",
			},
		},
	}
}

func (r *ForwardingResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultGlobal()
}

func (r *ForwardingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ForwardingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan forwardingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fwd, err := r.api.CreateForwarding(&mailboxsdk.CreateForwardingRequest{
		MailboxID: plan.MailboxID.ValueString(),
		Email:     plan.Email.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create mailbox forwarding", err.Error())

		return
	}

	fwd, err = waitForForwarding(ctx, r.api, fwd.ID, defaultMailboxTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for mailbox forwarding", err.Error())

		return
	}

	state := convertForwardingToState(fwd)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(fwd.ID))...)
}

func (r *ForwardingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    forwardingResourceModel
		identity forwardingResourceIdentityModel
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

	forwardingID := state.ID.ValueString()
	if identityAvailable {
		forwardingID = identity.ID.ValueString()
	}

	if forwardingID == "" {
		resp.Diagnostics.AddError("Forwarding ID not set", "Cannot refresh mailbox forwarding without an ID")

		return
	}

	fwd, err := r.api.GetForwarding(&mailboxsdk.GetForwardingRequest{ForwardingID: forwardingID}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to get mailbox forwarding", err.Error())

		return
	}

	state = convertForwardingToState(fwd)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(fwd.ID))...)
}

func (r *ForwardingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  forwardingResourceModel
		state forwardingResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Email.Equal(state.Email) {
		_, err := r.api.UpdateForwarding(&mailboxsdk.UpdateForwardingRequest{
			ForwardingID: plan.ID.ValueString(),
			Email:        providertypes.ExpandStringPtr(plan.Email.ValueString()),
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update mailbox forwarding", err.Error())

			return
		}

		fwd, err := waitForForwarding(ctx, r.api, plan.ID.ValueString(), defaultMailboxTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed waiting for mailbox forwarding after update", err.Error())

			return
		}

		newState := convertForwardingToState(fwd)
		resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
		resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(fwd.ID))...)

		return
	}

	fwd, err := r.api.GetForwarding(&mailboxsdk.GetForwardingRequest{ForwardingID: plan.ID.ValueString()}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get mailbox forwarding after update", err.Error())

		return
	}

	newState := convertForwardingToState(fwd)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(fwd.ID))...)
}

func (r *ForwardingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state forwardingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.api.DeleteForwarding(&mailboxsdk.DeleteForwardingRequest{ForwardingID: state.ID.ValueString()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete mailbox forwarding", err.Error())

		return
	}

	err = waitForForwardingDeleted(ctx, r.api, state.ID.ValueString(), defaultMailboxTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for mailbox forwarding deletion", err.Error())
	}
}

func (r *ForwardingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func convertForwardingToState(fwd *mailboxsdk.Forwarding) forwardingResourceModel {
	return forwardingResourceModel{
		ID:        types.StringValue(fwd.ID),
		MailboxID: types.StringValue(fwd.MailboxID),
		Email:     types.StringValue(fwd.Email),
		Status:    types.StringValue(fwd.Status.String()),
		CreatedAt: flattenTime(fwd.CreatedAt),
		UpdatedAt: flattenTime(fwd.UpdatedAt),
	}
}
