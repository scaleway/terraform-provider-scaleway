package partner

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	partner "github.com/scaleway/scaleway-sdk-go/api/partner/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ resource.Resource                = (*PartnerOrganizationResource)(nil)
	_ resource.ResourceWithConfigure   = (*PartnerOrganizationResource)(nil)
	_ resource.ResourceWithImportState = (*PartnerOrganizationResource)(nil)
	_ resource.ResourceWithIdentity    = (*PartnerOrganizationResource)(nil)
)

func NewPartnerOrganizationResource() resource.Resource {
	return &PartnerOrganizationResource{}
}

type PartnerOrganizationResource struct {
	partnerAPI *partner.API
	meta       *meta.Meta
}

type partnerOrganizationResourceModel struct {
	ID             types.String `tfsdk:"id"`
	PartnerID      types.String `tfsdk:"partner_id"`
	Email          types.String `tfsdk:"email"`
	Name           types.String `tfsdk:"organization_name"`
	OwnerFirstname types.String `tfsdk:"owner_firstname"`
	OwnerLastname  types.String `tfsdk:"owner_lastname"`
	PhoneNumber    types.String `tfsdk:"phone_number"`
	CustomerID     types.String `tfsdk:"customer_id"`
	Status         types.String `tfsdk:"status"`
	LockedBy       types.String `tfsdk:"locked_by"`
	LockReason     types.String `tfsdk:"lock_reason_message"`
	CreatedAt      types.String `tfsdk:"created_at"`
	LockedAt       types.String `tfsdk:"locked_at"`
}

type partnerOrganizationResourceIdentityModel = framework.GlobalIdentity

func (r *PartnerOrganizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_organization"
}

//go:embed descriptions/organization_resource.md
var partnerOrganizationResourceDescription string

func (r *PartnerOrganizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: partnerOrganizationResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the organization resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The email of the new organization owner.",
			},
			"organization_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the organization you want to create. Usually the company name.",
			},
			"partner_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Your personal partner_id. This is the same as your Organization ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owner_firstname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The first name of the new organization owner.",
			},
			"owner_lastname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The last name of the new organization owner.",
			},
			"phone_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The phone number of the new organization owner.",
			},
			"customer_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A custom ID for the customer in your own infrastructure.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the organization.",
			},
			"locked_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Originator of the lock (partner or scaleway).",
			},
			"lock_reason_message": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Human-readable reason if the organization is locked.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date of organization creation.",
			},
			"locked_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date of lock.",
			},
		},
	}
}

func (r *PartnerOrganizationResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultGlobal()
}

func (r *PartnerOrganizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.partnerAPI = partner.NewAPI(meta.ExtractScwClient(m))
}

func (r *PartnerOrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data partnerOrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	partnerOrganizationID, err := resolvePartnerOrganizationID(data.PartnerID.ValueString(), r.meta)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create partner organization",
			err.Error(),
		)

		return
	}

	createReq := &partner.CreateOrganizationRequest{
		PartnerID:        partnerOrganizationID,
		Email:            data.Email.ValueString(),
		OrganizationName: data.Name.ValueString(),
		OwnerFirstname:   data.OwnerFirstname.ValueString(),
		OwnerLastname:    data.OwnerLastname.ValueString(),
		CustomerID:       data.CustomerID.ValueString(),
	}

	if !data.PhoneNumber.IsNull() && !data.PhoneNumber.IsUnknown() {
		phoneNumber := data.PhoneNumber.ValueString()
		createReq.PhoneNumber = &phoneNumber
	}

	res, err := r.partnerAPI.CreateOrganization(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create partner organization",
			fmt.Sprintf("Failed to create partner organization: %s", err),
		)

		return
	}

	state := convertOrganizationToState(res, partnerOrganizationID, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(res.ID))...)
}

func (r *PartnerOrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    partnerOrganizationResourceModel
		identity partnerOrganizationResourceIdentityModel
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

	var organizationID string
	if identityAvailable {
		organizationID = locality.ExpandID(identity.ID.ValueString())
	} else {
		organizationID = locality.ExpandID(state.ID.ValueString())
	}

	res, err := r.partnerAPI.GetOrganization(&partner.GetOrganizationRequest{
		OrganizationID: organizationID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to read partner organization",
			fmt.Sprintf("Failed to read partner organization: %s", err),
		)

		return
	}

	partnerOrganizationID, err := resolvePartnerOrganizationID(state.PartnerID.ValueString(), r.meta)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read partner organization",
			err.Error(),
		)

		return
	}

	state = convertOrganizationToState(res, partnerOrganizationID, state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(res.ID))...)
}

func (r *PartnerOrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data partnerOrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := locality.ExpandID(data.ID.ValueString())

	updateReq := &partner.UpdateOrganizationRequest{
		OrganizationID: organizationID,
	}

	if !data.Email.IsNull() && !data.Email.IsUnknown() {
		email := data.Email.ValueString()
		updateReq.Email = &email
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		updateReq.Name = &name
	}

	if !data.OwnerFirstname.IsNull() && !data.OwnerFirstname.IsUnknown() {
		ownerFirstname := data.OwnerFirstname.ValueString()
		updateReq.OwnerFirstname = &ownerFirstname
	}

	if !data.OwnerLastname.IsNull() && !data.OwnerLastname.IsUnknown() {
		ownerLastname := data.OwnerLastname.ValueString()
		updateReq.OwnerLastname = &ownerLastname
	}

	if !data.PhoneNumber.IsNull() && !data.PhoneNumber.IsUnknown() {
		phoneNumber := data.PhoneNumber.ValueString()
		updateReq.PhoneNumber = &phoneNumber
	}

	if !data.CustomerID.IsNull() && !data.CustomerID.IsUnknown() {
		customerID := data.CustomerID.ValueString()
		updateReq.CustomerID = &customerID
	}

	res, err := r.partnerAPI.UpdateOrganization(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update partner organization",
			fmt.Sprintf("Failed to update partner organization: %s", err),
		)

		return
	}

	partnerOrganizationID, err := resolvePartnerOrganizationID(data.PartnerID.ValueString(), r.meta)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update partner organization",
			err.Error(),
		)

		return
	}

	state := convertOrganizationToState(res, partnerOrganizationID, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(res.ID))...)
}

func (r *PartnerOrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerOrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Partner organizations cannot be deleted via API, they can only be locked/unlocked
	// We simply remove the resource from state
	resp.State.RemoveResource(ctx)
}

func (r *PartnerOrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)

	if partnerOrganizationID, exists := r.meta.ScwClient().GetDefaultOrganizationID(); exists {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("partner_id"), partnerOrganizationID)...)
	}
}

func resolvePartnerOrganizationID(partnerID string, m *meta.Meta) (string, error) {
	partnerOrganizationID := locality.ExpandID(partnerID)
	if partnerOrganizationID != "" {
		return partnerOrganizationID, nil
	}

	defaultPartnerOrgID, exists := m.ScwClient().GetDefaultOrganizationID()
	if exists {
		return defaultPartnerOrgID, nil
	}

	return "", errors.New("either set partner_id or configure a default organization")
}

func convertOrganizationToState(organization *partner.Organization, partnerOrganizationID string, data partnerOrganizationResourceModel) partnerOrganizationResourceModel {
	data.ID = types.StringValue(organization.ID)
	data.PartnerID = types.StringValue(partnerOrganizationID)
	data.Email = types.StringValue(organization.Email)
	data.Name = types.StringValue(organization.Name)
	data.OwnerFirstname = types.StringValue(organization.OwnerFirstname)
	data.OwnerLastname = types.StringValue(organization.OwnerLastname)
	data.CustomerID = types.StringValue(organization.CustomerID)
	data.Status = types.StringValue(string(organization.Status))
	data.LockedBy = types.StringValue(string(organization.LockedBy))
	data.LockReason = types.StringValue(organization.LockReasonMessage)

	if organization.PhoneNumber != nil {
		data.PhoneNumber = types.StringValue(*organization.PhoneNumber)
	} else {
		data.PhoneNumber = types.StringNull()
	}

	if organization.CreatedAt != nil {
		data.CreatedAt = types.StringValue(organization.CreatedAt.String())
	}

	if organization.LockedAt != nil {
		data.LockedAt = types.StringValue(organization.LockedAt.String())
	}

	return data
}
