package partner

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	partner "github.com/scaleway/scaleway-sdk-go/api/partner/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*PartnerOrganizationRequestAdminRoleAction)(nil)
	_ action.ActionWithConfigure = (*PartnerOrganizationRequestAdminRoleAction)(nil)
)

type PartnerOrganizationRequestAdminRoleAction struct {
	partnerAPI *partner.API
}

func (a *PartnerOrganizationRequestAdminRoleAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	a.partnerAPI = partner.NewAPI(meta.ExtractScwClient(m))
}

func (a *PartnerOrganizationRequestAdminRoleAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_organization_request_admin_role"
}

type PartnerOrganizationRequestAdminRoleActionModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	Username       types.String `tfsdk:"username"`
	Email          types.String `tfsdk:"email"`
	Password       types.String `tfsdk:"password"`
}

func NewPartnerOrganizationRequestAdminRoleAction() action.Action {
	return &PartnerOrganizationRequestAdminRoleAction{}
}

//go:embed descriptions/organization_request_admin_role_action.md
var partnerOrganizationRequestAdminRoleActionDescription string

func (a *PartnerOrganizationRequestAdminRoleAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: partnerOrganizationRequestAdminRoleActionDescription,
		Description:         partnerOrganizationRequestAdminRoleActionDescription,
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the organization to request admin role for.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The member username.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The member email.",
				Validators: []validator.String{
					verify.IsStringEmail(),
				},
			},
			"password": schema.StringAttribute{
				Required:    true,
				Description: "The member password.",
			},
		},
	}
}

func (a *PartnerOrganizationRequestAdminRoleAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data PartnerOrganizationRequestAdminRoleActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.partnerAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured partnerAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.OrganizationID.IsNull() || data.OrganizationID.IsUnknown() || data.OrganizationID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing organization_id",
			"The organization_id attribute is required to request admin role.",
		)

		return
	}

	orgID := locality.ExpandID(data.OrganizationID.ValueString())

	if data.Username.IsNull() || data.Username.IsUnknown() || data.Username.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing username",
			"The username attribute is required to request admin role.",
		)

		return
	}

	if data.Email.IsNull() || data.Email.IsUnknown() || data.Email.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing email",
			"The email attribute is required to request admin role.",
		)

		return
	}

	if data.Password.IsNull() || data.Password.IsUnknown() || data.Password.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing password",
			"The password attribute is required to request admin role.",
		)

		return
	}

	err := a.partnerAPI.RequestAdminRole(&partner.RequestAdminRoleRequest{
		OrganizationID: orgID,
		Username:       data.Username.ValueString(),
		Email:          data.Email.ValueString(),
		Password:       data.Password.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Partner RequestAdminRole action",
			fmt.Sprintf("Failed to request admin role for organization %s: %s", orgID, err),
		)

		return
	}
}
