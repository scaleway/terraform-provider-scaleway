package partner

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	partner "github.com/scaleway/scaleway-sdk-go/api/partner/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ list.ListResource              = (*PartnerOrganizationListResource)(nil)
	_ list.ListResourceWithConfigure = (*PartnerOrganizationListResource)(nil)
)

type PartnerOrganizationListResource struct {
	meta       *meta.Meta
	partnerAPI *partner.API
}

func (r *PartnerOrganizationListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.partnerAPI = partner.NewAPI(meta.ExtractScwClient(m))
}

func NewPartnerOrganizationListResource() list.ListResource {
	return &PartnerOrganizationListResource{}
}

func (r *PartnerOrganizationListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_organization"
}

func (r *PartnerOrganizationListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Description: "List organizations managed by the partner.",
		Attributes: map[string]schema.Attribute{
			"order_by": schema.StringAttribute{
				Description: "Order by field. Can be `created_at_asc` or `created_at_desc`.",
				Optional:    true,
				Validators: []validator.String{
					verify.ValidateEnumFramework[partner.ListOrganizationsRequestOrderBy](),
				},
			},
			"status": schema.StringAttribute{
				Description: "Filter by status. Can be `unknown_status`, `opened`, `locked` or `closed`.",
				Optional:    true,
				Validators: []validator.String{
					verify.ValidateEnumFramework[partner.OrganizationStatus](),
				},
			},
			"email": schema.StringAttribute{
				Description: "Filter by email",
				Optional:    true,
			},
			"customer_id": schema.StringAttribute{
				Description: "Filter by customer ID",
				Optional:    true,
			},
			"partner_id": schema.StringAttribute{
				Description: "Your personal partner_id. This is the same as your Organization ID.",
				Optional:    true,
			},
			"locked_by": schema.StringAttribute{
				Description: "Filter by `locked_by`. Can be `unknown_locked_by`, `partner` or `scaleway`.",
				Optional:    true,
				Validators: []validator.String{
					verify.ValidateEnumFramework[partner.OrganizationLockedBy](),
				},
			},
		},
	}
}

type PartnerOrganizationListResourceModel struct {
	OrderBy    types.String `tfsdk:"order_by"`
	Status     types.String `tfsdk:"status"`
	Email      types.String `tfsdk:"email"`
	CustomerID types.String `tfsdk:"customer_id"`
	PartnerID  types.String `tfsdk:"partner_id"`
	LockedBy   types.String `tfsdk:"locked_by"`
}

func (r *PartnerOrganizationListResource) FetchOrganizations(ctx context.Context, data PartnerOrganizationListResourceModel) ([]*partner.Organization, error) {
	request := &partner.ListOrganizationsRequest{}

	if !data.OrderBy.IsNull() && !data.OrderBy.IsUnknown() {
		request.OrderBy = partner.ListOrganizationsRequestOrderBy(data.OrderBy.ValueString())
	}

	if !data.Status.IsNull() && !data.Status.IsUnknown() {
		request.Status = partner.OrganizationStatus(data.Status.ValueString())
	}

	if !data.Email.IsNull() && !data.Email.IsUnknown() {
		email := strings.TrimSpace(data.Email.ValueString())
		if email != "" {
			request.Email = &email
		}
	}

	if !data.CustomerID.IsNull() && !data.CustomerID.IsUnknown() {
		customerID := strings.TrimSpace(data.CustomerID.ValueString())
		if customerID != "" {
			request.CustomerID = &customerID
		}
	}

	if !data.LockedBy.IsNull() && !data.LockedBy.IsUnknown() {
		request.LockedBy = partner.OrganizationLockedBy(data.LockedBy.ValueString())
	}

	response, err := r.partnerAPI.ListOrganizations(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Organizations, nil
}

func (r *PartnerOrganizationListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data PartnerOrganizationListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allOrganizations, err := r.FetchOrganizations(ctx, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Partner Organizations", "Failed to list Partner Organizations: "+err.Error()),
		})

		return
	}

	partnerOrganizationID, err := resolvePartnerOrganizationID(data.PartnerID.ValueString(), r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Partner Organizations", err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, organization := range allOrganizations {
			result := req.NewListResult(ctx)
			result.DisplayName = organization.Name

			diags := result.Identity.Set(ctx, framework.SetGlobalIdentity(organization.ID))
			result.Diagnostics.Append(diags...)

			if req.IncludeResource {
				resourceModel := convertOrganizationToState(organization, partnerOrganizationID, partnerOrganizationResourceModel{})
				diags := result.Resource.Set(ctx, resourceModel)
				result.Diagnostics.Append(diags...)
			}

			if !push(result) {
				return
			}
		}
	}
}
